package opennebula

import (
	"fmt"
	"github.com/kolo/xmlrpc"
	"log"
	"strconv"
)

type Client struct {
	Rcp      xmlrpc.Client
	session  string
	Username string
	Password string
}

func NewClient(endpoint, username, password string) (*Client, error) {
	client, err := xmlrpc.NewClient(endpoint, nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		Rcp:      *client,
		session:  fmt.Sprintf("%s:%s", username, password),
		Username: username,
		Password: password,
	}, nil
}

func (c *Client) Call(command string, args ...interface{}) (string, error) {
	var result []interface{}

	args = append([]interface{}{c.session}, args...)

	if err := c.Rcp.Call(command, args, &result); err != nil {
		return "", err
	}

	res, err := c.IsSuccess(result)
	if err != nil {
		return "", err
	}

	return res, nil
}

func (c *Client) IsSuccess(result []interface{}) (res string, err error) {
	if !result[0].(bool) {
		err = fmt.Errorf("%s", result[1].(string))
		return
	}

	if w, ok := result[1].(int64); ok {
		res = strconv.FormatInt(w, 10)
	} else if w, ok := result[1].(string); ok {
		res = w
	}

	return
}

func intId(id string) int {
	i, err := strconv.Atoi(id)
	if err != nil {
		log.Fatalf("Unexpected ID %s received from OpenNebula. Expected an integer", id)
	}

	return i
}
