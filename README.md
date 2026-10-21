# Net-Cat

A terminal-based group chat application written in Go. Net-Cat uses TCP connections to let multiple clients exchange messages through one server in real time.

The server welcomes each client with an ASCII Linux logo, asks for a username, keeps message history, and notifies connected clients when someone joins, leaves, or changes their name.

## Features

- Start the TCP server on the default port `8989` or choose a custom port.
- Support up to **10 simultaneous connections**.
- Display an ASCII Linux logo and ask every new client to enter a username.
- Reject empty or duplicate usernames without disconnecting the client.
- Compare usernames without case sensitivity (`Ali` and `ali` are considered the same name).
- Allow connected clients to change their name using `/name NewName`.
- Notify other clients when someone joins, leaves, or changes their name.
- Exchange messages between all connected clients, including the sender.
- Attach the sender's name and the server's timestamp to every chat message.
- Save previous messages in memory and show them to new clients when they join.
- Ignore messages containing only spaces or an empty line.
- Handle multiple connections using goroutines and protect shared data using mutexes.
- Include automated tests for connection limits, usernames, and basic server behavior.

## Technology Stack

| Technology | Purpose |
| --- | --- |
| Go | TCP server, client handling, chat logic, and concurrency |
| Go standard library | Networking, input processing, synchronization, and timestamps |
| TCP | Reliable, connection-oriented communication between server and clients |
| Netcat (`nc`) | Terminal client used to join and test the chat |
| Goroutines | Handle multiple connected clients concurrently |
| Mutexes (`sync.Mutex`) | Coordinate access to shared server data and client writes |
| In-memory slice and map | Store chat history and track active clients |

No external Go packages, database, frontend, or web browser are required.

## Project Structure

```text
net-cat/
├── go.mod
├── main.go
├── README.md
├── .gitignore
└── tcpchat/
    ├── server.go
    ├── client.go
    ├── logo.go
    └── server_test.go
```

| File | Responsibility |
| --- | --- |
| `main.go` | Reads command-line arguments and starts the server |
| `tcpchat/server.go` | Accepts connections, enforces connection limits, tracks clients, saves history, broadcasts messages, and changes usernames |
| `tcpchat/client.go` | Welcomes clients, validates usernames, reads messages, processes `/name`, and handles disconnections |
| `tcpchat/logo.go` | Stores the welcome banner and username prompt |
| `tcpchat/server_test.go` | Tests connection limits, basic server state, and username behavior |
| `go.mod` | Defines the project module and Go version |
| `.gitignore` | Excludes generated executables and other local files from version control |

## Requirements

- Go compatible with the version declared in `go.mod` (**Go 1.25.6**).
- Ubuntu/Linux, Windows, or another system that can run Go.
- A terminal with Netcat (`nc`) or Ncat for connecting as a chat client.
- An available TCP port (`8989` by default).

For Ubuntu, install Netcat if necessary:

```bash
sudo apt update
sudo apt install -y netcat-openbsd
```

## Build and Run

Run the commands from the project root directory.

**Build the Linux executable:**

```bash
cd net-cat
go build -o TCPChat .
```

**Start on the default port:**

```bash
./TCPChat
```

Expected output:

```text
Listening on the port :8989
```

**Start on a custom port:**

```bash
./TCPChat 2525
```

Expected output:

```text
Listening on the port :2525
```

**Run without building an executable first:**

```bash
go run .
```

**Invalid command-line arguments:**

```bash
./TCPChat 2525 localhost
```

Output:

```text
[USAGE]: ./TCPChat $port
```

The server continues running while waiting for clients. Press `Ctrl+C` in the server terminal to stop it.

**Windows:** Build with `go build -o TCPChat.exe .` and launch `./TCPChat.exe` from Git Bash or PowerShell. Linux executables and Windows executables are different, so build for the operating system where the server will run.

## Connecting Clients

Keep the server running in the first terminal. Open a second terminal and connect using Netcat:

```bash
nc 127.0.0.1 8989
```

Open a third terminal and repeat the same command to connect another client.

If the server is running on port `2525`, connect using:

```bash
nc 127.0.0.1 2525
```

The client first sees the welcome banner:

```text
Welcome to TCP-Chat!
         _nnnn_
        dGGGGMMb
       @p~qp~~qMb
       ...
[ENTER YOUR NAME]:
```

Enter a name and press Enter to join the chat.

### Example Conversation

Client 1 joins as `Ali`. Client 2 joins as `Ahmed`.

**Ali's terminal:**

```text
[ENTER YOUR NAME]: Ali
[2026-10-21 20:10:00][Ali]:
Ahmed has joined our chat...
[2026-10-21 20:10:05][Ali]:Hello Ahmed

[2026-10-21 20:10:08][Ali]:Hello Ahmed
[2026-10-21 20:10:08][Ali]:
```

**Ahmed's terminal receives:**

```text
[2026-10-21 20:10:08][Ali]:Hello Ahmed
[2026-10-21 20:10:08][Ahmed]:
```

Timestamps are examples. The server uses the time when each message is processed. Because Netcat can display locally typed input, the sender may also see an additional server-formatted copy of their message.

## Application Architecture

The application consists of one TCP server and multiple terminal clients.

```text
                  +------------------+
Client Ali  <---->|                  |
Client Ahmed<---->|  Go TCP Server   |----> Message History
Client Waleed<-->|                  |
                  +------------------+
```

Each client establishes a separate TCP connection to the same server. The server handles incoming lines, manages usernames, and sends messages to connected clients.

## How the Chat Works

### Server Startup

1. `main.go` selects port `8989` unless a custom port is provided.
2. `RunServer()` starts a TCP listener on that port.
3. The server waits for new connections using `Accept()`.
4. A connection slot is reserved if fewer than 10 connections are active.
5. A separate goroutine begins handling each accepted client.

### Client Connection and Username

1. The server sends the welcome banner and asks for a username.
2. An empty username is rejected.
3. A username already used by another connected client is rejected, regardless of capitalization.
4. The client can try another username without reconnecting.
5. Once an available name is chosen, the server sends previous messages, registers the client, and notifies the others.

### Sending Messages

1. A client types a line and presses Enter.
2. The server ignores the line if it is blank or contains only spaces.
3. Otherwise, the server formats it as `[YYYY-MM-DD HH:MM:SS][name]:message`.
4. The message is saved in history and delivered to all connected clients.
5. A new client joining later receives the saved history in order.

### Disconnecting

When a client closes Netcat (for example, using `Ctrl+C`), the server removes that client, informs the remaining members, and makes the connection slot and username available again.

## Commands and Username Rules

| Input | Result |
| --- | --- |
| `Ali` at the name prompt | Joins the chat if the name is available |
| `ali` when `Ali` is connected | Rejected as a duplicate name |
| Empty name | Asked to enter a name again |
| `Hello everyone` | Sends a timestamped chat message |
| Empty chat message | Ignored |
| `/name Ahmed` | Changes the current username if `Ahmed` is available |
| `/name` | Shows command usage |

**Change-name example:**

Ali types:

```text
/name Ahmed
```

Ali sees:

```text
Your name is now Ahmed.
```

Other clients see:

```text
Ali changed their name to Ahmed...
```

A duplicate new name is rejected and the client's existing name is preserved. Commands such as `/name` are not saved as ordinary chat messages.

## Connection Limit

The server allows up to **10 simultaneous TCP connections**, including connections where a user has not entered a name yet.

If an 11th client tries to connect, the server sends:

```text
Chat is full. Maximum 10 connections allowed.
```

The extra connection is then closed; existing clients remain connected. When someone leaves, another client can use the available slot.

## Concurrency and Synchronization

The server uses **goroutines** so one client's activity does not prevent the program from accepting other connections.

Two mutexes coordinate concurrent operations:

| Mutex | Responsibility |
| --- | --- |
| `Server.mu` | Protects the connection count, active-client list, usernames, and shared message history |
| `Client.writeMu` | Prevents simultaneous writes to the same client's TCP connection |

Network writes use a timeout to limit how long an unresponsive connection can delay a send. Messages and shared data are stored in memory while the server runs.

## Error Handling

| Situation | Behavior |
| --- | --- |
| Server port already in use | Startup fails and prints a server error |
| Too many command-line arguments | Prints `[USAGE]: ./TCPChat $port` |
| More than 10 TCP connections | Refuses the additional connection |
| Empty username | Prompts for another name |
| Duplicate username | Asks the client to choose another name |
| Empty chat message | Does not broadcast it |
| Username change to a taken name | Keeps the existing username and shows an error |
| Client disconnects | Closes the connection, frees the slot, and notifies other clients |
| A client cannot receive a message | A send error may cause its connection to be closed |

## Manual Testing

Start the server before testing. Keep the existing clients connected unless a test explicitly tells you to disconnect them.

| Test | Expected Result |
| --- | --- |
| Run `./TCPChat` | Server listens on port `8989` |
| Run `./TCPChat 2525` | Server listens on port `2525` |
| Run `./TCPChat 2525 localhost` | Usage message appears |
| Connect two clients using `nc` | Both receive the welcome banner and can join |
| Join as `Ali` and then `ali` | Second client must choose a different name |
| Join three clients | Existing clients receive join notifications |
| Send a message | Connected clients receive the timestamped message |
| Connect a new client after sending messages | New client receives previous messages |
| Disconnect one client | Other clients remain connected and receive a leave notification |
| Connect 10 clients, then an 11th | The 11th is rejected |
| Disconnect one of the 10 clients | A new connection is accepted |
| Enter `/name NewName` | Name changes and other clients are notified |
| Rename to an existing name | Rename is rejected |

### Automated Tests

Run in the project root:

```bash
go fmt ./...
go test ./...
go test -race ./...
go vet ./...
```

The tests cover the connection limit, initialization, duplicate names, basic rename behavior, and simultaneous connection reservations. Passing tests and the race detector does **not** prove that every network or concurrency situation is handled perfectly; also perform the manual tests above.

## Connecting from Another Computer

Clients on another computer can connect using the machine running the server's reachable IP address:

```bash
nc <server-ip> 8989
```

For example, replace `<server-ip>` with the IP of the actual machine hosting the server. The server's port must be reachable, and the network must permit device-to-device connections.

**Important for WSL 2 users:** When the server runs inside Ubuntu WSL, other computers might not be able to reach it directly through the Windows Wi-Fi address. Additional Windows/WSL network configuration may be required. Alternatively, run the Go server directly on Windows to avoid WSL port forwarding. In either case, Windows Firewall and network rules may still apply. Do not disable the entire firewall just to test the chat.

## Current Limitations

- Chat history is stored only in memory and disappears when the server stops.
- The project has one shared chat room; separate rooms are not implemented.
- There is no login system, password authentication, or message encryption at the application level.
- Netcat may show the sender's typed message as well as the server-formatted copy.
- A very slow client can delay other operations briefly because some outgoing messages are sent while the server holds its shared mutex.
- The application is designed for learning and local-network testing, not deployment as a production public chat service.

## Learning Outcomes

This project demonstrates:

- Creating and listening on a TCP server with Go.
- Accepting and handling multiple client connections.
- Using goroutines for concurrent activity.
- Using mutexes to coordinate access to shared data.
- Reading terminal input with `bufio.Scanner`.
- Storing clients in maps and chat history in slices.
- Validating usernames and processing simple commands.
- Formatting messages with timestamps.
- Broadcasting messages and detecting disconnections.
- Writing and running Go unit tests.
- Organizing Go code into separate files and packages.

## GitHub and Reboot Audit

Generated executables such as `TCPChat` and `TCPChat.exe` are excluded by `.gitignore`; rebuild them on the operating system where they will be used.

To save and push changes:

```bash
git status
git add .
git commit -m "Document Net-Cat and add unique usernames and rename support"
git push origin main
```

The Reboot audit evaluates the commit selected for that audit. Pushing a new commit does not change the commit used by an already-created audit automatically.
