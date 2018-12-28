package flooding

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"
)

var eod = []byte("\n")

func writeToConnection(conn *net.TCPConn, json string) (int, error) {
	n, werr := fmt.Fprintf(conn, json+"\n")
	return n, werr
}

func readFromConnection(conn *net.TCPConn) (string, error) {
	json, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	return json, nil
}

// GetNewUnusedPort generate a random local ip port
func GetNewUnusedPort() string {
	// TODO: you should check whether this random port is really unused
	// assuming IPs are all localhost
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	// generate a random port between [min, max]
	min := 3000
	max := 60000
	port := random.Intn(max-min) + min
	return strconv.Itoa(port)
}
