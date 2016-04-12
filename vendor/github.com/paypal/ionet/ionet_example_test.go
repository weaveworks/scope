package ionet

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

// ExampleListener uses ionet to start a http server,
// connects to it with byte buffers for readers/writers,
// makes a request, and receives and parses the response.
func ExampleListener() {
	// Create an ionet.Listener
	l := new(Listener)

	// Set up an http server that handles requests using that listener
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("Hello!"))
	}))
	server := &http.Server{Handler: mux}
	go server.Serve(l)

	// Dial our listener with an http request; write the response to a buffer
	// The response buffer is called w, as in writer. That is intentional;
	// all ionet variables are named from the server's perspective, and the
	// server writes into the response buffer. See the Conn documentation.
	r := bytes.NewBufferString("GET / HTTP/1.1\r\nConnection: close\r\n\r\n")
	w := new(bytes.Buffer)
	conn, err := l.Dial(r, w)
	if err != nil {
		fmt.Printf("Dial error: %v\n", err)
		return
	}

	// Wait for the connection to close
	conn.Wait()

	// Parse the response (stored in w)
	buf := bufio.NewReader(w)
	resp, err := http.ReadResponse(buf, new(http.Request))
	if err != nil {
		fmt.Printf("Response parse error: %v\n", err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Response read error: %v\n", err)
		return
	}
	resp.Body.Close()

	// Display the response
	fmt.Println(string(body))

	// Output:
	// Hello!
}
