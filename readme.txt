ZomCleaner is a simple "reaper", it is a lightweight web-based dashboard designed to monitor and manage specific low-level system resources and help in optimisation.
Clean0.go interacts with the "/proc" file system and triggers *kernel-level* actions, thus is designed specifically for compatibility with linux. 
Clean1.go uses powershell commands to work on windows. **note**- Powershell needs to be run as administrator for full use of the program. [In the same folder as clean1.go yhe user should run setup.ps1 with ./setup.ps1, to enable running scripts use the follwing command "Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser"]
.
One can run the go file as "go run Clean0.go"; to start the server.
.
The UI interface is served on localhost port 8080 once the server is active. http://127.0.0.1:8080
