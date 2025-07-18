# 8paste
8paste is a simple paste service which supports:
- Uploading/downloading paste as file or text
- Infinite/temporary pastes (partially incomplete)
- Password protected pastes
- Command-line client and a frontend
# Usage
8paste has two components: client and server
### Server usage
```
server -p <PORT>
```
### Client usage
```
-g <ID>: Get paste with the specified ID
-u <content>: Upload paste with the specified content
-f <path>: Upload paste as a file from the specified path
...
To get more information about the arguments use the -h flag
```
# Building
To build the client or the server locate to the project root and run the specified command for go to install the dependencies
```bash
go mod download
```
After this command is done locate to the component directory you want to build (server or client) and then run the command
```bash
go build
```
