package agent

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

// IMAP-выгрузка почтовых ящиков в .eml. Только TLS (порт 993); учётка из
// credentials (username + секрет), курсор инкремента — дата в source.LastSync.

type imapMessage struct {
	UID          uint32
	InternalDate time.Time
	Body         func() (io.ReadCloser, error)
}

func imapClientNew(conn net.Conn) (*client.Client, error) {
	c, err := client.New(conn)
	if err != nil {
		return nil, fmt.Errorf("imap handshake: %w", err)
	}
	return c, nil
}

func imapFolders(c *client.Client) ([]string, error) {
	ch := make(chan *imap.MailboxInfo, 20)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", ch)
	}()
	var folders []string
	for m := range ch {
		folders = append(folders, m.Name)
	}
	if err := <-done; err != nil {
		return folders, err
	}
	return folders, nil
}

// imapFetchSince забирает письма новее since из папки folder. Тело
// открывается лениво: FETCH выполняется по одному письму при чтении архива.
func imapFetchSince(c *client.Client, folder string, since time.Time) ([]imapMessage, error) {
	mbox, err := c.Select(folder, true)
	if err != nil {
		return nil, fmt.Errorf("select %s: %w", folder, err)
	}
	if mbox.Messages == 0 {
		return nil, nil
	}

	criteria := imap.NewSearchCriteria()
	if !since.IsZero() {
		criteria.Since = since
	}
	uids, err := c.UidSearch(criteria)
	if err != nil {
		// Сервер без SEARCH SINCE — выгружаем весь ящик (пустой критерий = ALL).
		if uids, err = c.UidSearch(imap.NewSearchCriteria()); err != nil {
			return nil, fmt.Errorf("search %s: %w", folder, err)
		}
	}
	var out []imapMessage
	for _, uid := range uids {
		out = append(out, imapMessage{
			UID: uid,
			// InternalDate достаём в ленивом FETCH ниже; по умолчанию — now.
			InternalDate: time.Now(),
			Body: func() (io.ReadCloser, error) {
				seqSet := new(imap.SeqSet)
				seqSet.AddNum(uid)
				section := &imap.BodySectionName{}
				messages := make(chan *imap.Message, 1)
				done := make(chan error, 1)
				go func() {
					done <- c.UidFetch(seqSet, []imap.FetchItem{section.FetchItem()}, messages)
				}()
				msg := <-messages
				err := <-done
				if msg == nil || err != nil {
					if err != nil {
						return nil, err
					}
					return nil, fmt.Errorf("uid %d not fetched", uid)
				}
				r := msg.GetBody(section)
				if r == nil {
					return nil, fmt.Errorf("uid %d: empty body", uid)
				}
				if d := msg.InternalDate; !d.IsZero() {
					for i := range out {
						if out[i].UID == uid {
							out[i].InternalDate = d
							break
						}
					}
				}
				return io.NopCloser(r), nil
			},
		})
	}
	return out, nil
}
