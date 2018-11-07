package email

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Message creates a email to be sent
type Message struct {
	To      string
	From    string
	Subject string
	Body    string
}

var (
	ports   = []int{25, 2525, 587}
	dialCtx context.Context
)

// SendContext : Set context
func SendContext(ctx context.Context) {
	dialCtx = ctx
}

// Send sends a message to recipient(s) listed in the 'To' field of a Message
func (m Message) Send() error {
	if !strings.Contains(m.To, "@") {
		return fmt.Errorf("Invalid recipient address: <%s>", m.To)
	}

	host := strings.Split(m.To, "@")[1]
	addrs, err := net.LookupMX(host)
	if err != nil {
		return err
	}

	if dialCtx == nil {
		// ctx = context.Background() // Cause long wait time
		ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
		defer cancel()
		dialCtx = ctx
	}

	c, err := newClient(addrs, ports)
	if err != nil {
		return err
	}

	err = send(m, c)
	if err != nil {
		return err
	}

	return nil
}

func dialTimeout(addr string) (*smtp.Client, error) {
	var d net.Dialer

	conn, err := d.DialContext(dialCtx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(addr)

	return smtp.NewClient(conn, host)
}

func newClient(mx []*net.MX, ports []int) (*smtp.Client, error) {
	for i := range mx {
		for j := range ports {
			server := strings.TrimSuffix(mx[i].Host, ".")
			hostPort := fmt.Sprintf("%s:%d", server, ports[j])
			client, err := dialTimeout(hostPort)
			if err != nil {
				if j == len(ports)-1 {
					return nil, err
				}

				continue
			}

			return client, nil
		}
	}

	// Remove punctuation - error strings should not be capitalized or end with punctuation
	return nil, fmt.Errorf("Couldn't connect to servers %v on any common port", mx)
}

func send(m Message, c *smtp.Client) error {
	if err := c.Mail(m.From); err != nil {
		return err
	}

	if err := c.Rcpt(m.To); err != nil {
		return err
	}

	msg, err := c.Data()
	if err != nil {
		return err
	}

	if m.Subject != "" {
		_, err = msg.Write([]byte("Subject: " + m.Subject + "\r\n"))
		if err != nil {
			return err
		}
	}

	if m.From != "" {
		_, err = msg.Write([]byte("From: <" + m.From + ">\r\n"))
		if err != nil {
			return err
		}
	}

	if m.To != "" {
		_, err = msg.Write([]byte("To: <" + m.To + ">\r\n"))
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprint(msg, m.Body)
	if err != nil {
		return err
	}

	err = msg.Close()
	if err != nil {
		return err
	}

	err = c.Quit()
	if err != nil {
		return err
	}

	return nil
}
