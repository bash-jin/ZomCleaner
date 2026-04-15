# ZomCleaner Technical Documentation

ZomCleaner is a lightweight, web-based "reaper" designed to monitor and optimize low-level system resources.

## Deployment Instructions

### 1. Linux Environment (`Clean0.go`)
Designed for direct kernel-level interaction via the `/proc` filesystem.
* **Action:** Open terminal and execute:
    `go run Clean0.go`

### 2. Windows Environment (`Clean1.go`)
Uses PowerShell integration for system management. 
* **Prerequisite:** Launch PowerShell as **Administrator**.
* **Configuration:** Enable script execution by running:
    `Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser`
* **Initialization:** Within the project folder, execute:
    `./setup.ps1`
* **Action:** Start the server:
    `go run Clean1.go`

## Accessing the Interface
Once the server is active, the optimization dashboard is accessible via any web browser.
* **Local Address:** [http://127.0.0.1:8080](http://127.0.0.1:8080)

## Operational Requirements
* **Privileges:** Administrative/Root access is mandatory for full functionality.
* **Compatibility:** Ensure you run the specific file (`Clean0` for Linux, `Clean1` for Windows) corresponding to your host operating system.
