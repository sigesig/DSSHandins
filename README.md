	fmt.Println("Please enter peer IP and port (in the form peerIP:port) to connect")
	fmt.Scanln(&peerAddress)
	d := net.Dialer{Timeout: 60 * time.Second}
	conn, err := d.Dial("tcp", peerAddress)