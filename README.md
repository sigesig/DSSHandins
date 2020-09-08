fmt.Scanln(&peerAddress)
d := net.Dialer{Timeout: 60 * time.Second}