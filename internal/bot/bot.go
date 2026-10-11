package bot

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ZaViBiS/dont-worry-about-this-repo/internal/config"
	"ZaViBiS/dont-worry-about-this-repo/internal/db"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type update struct {
	Envelope struct {
		Source      string `json:"source"`
		DataMessage *struct {
			Message string `json:"message"`
		} `json:"dataMessage"`
	} `json:"envelope"`
}

func send(api, number, to, text string, base64Attachments ...string) error {
	payload := map[string]any{
		"message":    text,
		"number":     number,
		"recipients": []string{to},
	}
	if len(base64Attachments) > 0 {
		payload["base64_attachments"] = base64Attachments
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal send payload: %w", err)
	}

	resp, err := http.Post("http://"+api+"/v2/send", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("send status code: %d", resp.StatusCode)
	}

	return nil
}

// Run підключається до Signal API та обробляє вхідні повідомлення.
// Внутрішній цикл автоматично перепідключається при обриві зв'язку з логуванням помилок.
// Повертає помилку лише при скасуванні контексту або неможливості продовжити роботу.
func Run(ctx context.Context, database *sql.DB, cfg config.Config) error {
	berlinLoc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return fmt.Errorf("load Europe/Berlin timezone: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		wsURL := "ws://" + cfg.SignalAPI + "/v1/receive/" + cfg.SignalPhoneNumber
		log.Info().Str("url", wsURL).Msg("connecting to Signal websocket")

		conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
		if err != nil {
			log.Error().Err(err).Msg("websocket dial error, retrying in 5s")
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
				continue
			}
		}

		log.Info().Msg("connected to Signal websocket, listening for updates")

		readDone := make(chan error, 1)
		go func() {
			for {
				var u update
				if err := conn.ReadJSON(&u); err != nil {
					readDone <- fmt.Errorf("read json: %w", err)
					return
				}
				if m := u.Envelope.DataMessage; m != nil && m.Message != "" {
					log.Info().Str("from", u.Envelope.Source).Str("msg", m.Message).Msg("received message")

					trimmedMsg := strings.TrimSpace(m.Message)
					normalized := strings.TrimPrefix(trimmedMsg, "/")
					fields := strings.Fields(normalized)
					var command string
					if len(fields) > 0 {
						command = strings.ToLower(fields[0])
					}

					switch command {
					case "stat":
						records, err := db.GetAll(ctx, database)
						if err != nil {
							log.Error().Err(err).Msg("failed to get records for stat")
							if sendErr := send(cfg.SignalAPI, cfg.SignalPhoneNumber, u.Envelope.Source, "Помилка отримання даних з бази"); sendErr != nil {
								log.Error().Err(sendErr).Msg("failed to send reply")
							}
							continue
						}
						if len(records) == 0 {
							if err := send(cfg.SignalAPI, cfg.SignalPhoneNumber, u.Envelope.Source, "Немає записів для побудови статистики."); err != nil {
								log.Error().Err(err).Msg("failed to send reply")
							}
							continue
						}

						chartPNG, err := generateChart(records, berlinLoc)
						if err != nil {
							log.Error().Err(err).Msg("failed to generate chart")
							if sendErr := send(cfg.SignalAPI, cfg.SignalPhoneNumber, u.Envelope.Source, "Помилка побудови графіка"); sendErr != nil {
								log.Error().Err(sendErr).Msg("failed to send reply")
							}
							continue
						}

						encodedChart := "data:image/png;base64," + base64.StdEncoding.EncodeToString(chartPNG)
						caption := fmt.Sprintf("Статистика (всього записів: %d)", len(records))
						if err := send(cfg.SignalAPI, cfg.SignalPhoneNumber, u.Envelope.Source, caption, encodedChart); err != nil {
							log.Error().Err(err).Str("to", u.Envelope.Source).Msg("failed to send chart reply")
						}
					case "del":
						if len(fields) != 2 {
							if err := send(cfg.SignalAPI, cfg.SignalPhoneNumber, u.Envelope.Source, "вкажіть id запису: 'del <id>'"); err != nil {
								log.Error().Err(err).Str("to", u.Envelope.Source).Msg("failed to send reply")
							}
							continue
						}

						id, err := strconv.ParseInt(fields[1], 10, 64)
						if err != nil {
							if sendErr := send(cfg.SignalAPI, cfg.SignalPhoneNumber, u.Envelope.Source, "некоректний id запису: очікується число"); sendErr != nil {
								log.Error().Err(sendErr).Str("to", u.Envelope.Source).Msg("failed to send reply")
							}
							continue
						}

						deleted, err := db.Delete(ctx, database, id)
						if err != nil {
							log.Error().Err(err).Int64("id", id).Msg("failed to delete record")
							if sendErr := send(cfg.SignalAPI, cfg.SignalPhoneNumber, u.Envelope.Source, "помилка видалення запису"); sendErr != nil {
								log.Error().Err(sendErr).Str("to", u.Envelope.Source).Msg("failed to send reply")
							}
							continue
						}

						var reply string
						if deleted {
							reply = fmt.Sprintf("запис %d видалено", id)
						} else {
							reply = fmt.Sprintf("запис %d не знайдено", id)
						}
						if err := send(cfg.SignalAPI, cfg.SignalPhoneNumber, u.Envelope.Source, reply); err != nil {
							log.Error().Err(err).Str("to", u.Envelope.Source).Msg("failed to send reply")
						}
					default:
						record, err := db.Add(ctx, database)
						if err != nil {
							log.Error().Err(err).Msg("failed to insert record into db")
						}
						berlinTime := record.Timestamp.In(berlinLoc).Format("2006-01-02 15:04:05")
						text := fmt.Sprintf("запис: %d, час: %s", record.ID, berlinTime)
						if err := send(cfg.SignalAPI, cfg.SignalPhoneNumber, u.Envelope.Source, text); err != nil {
							log.Error().Err(err).Str("to", u.Envelope.Source).Msg("failed to send reply")
						}
					}

				}
			}
		}()

		select {
		case <-ctx.Done():
			_ = conn.Close()
			return ctx.Err()
		case err := <-readDone:
			_ = conn.Close()
			log.Warn().Err(err).Msg("websocket connection closed, reconnecting in 5s")
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
			}
		}
	}
}
