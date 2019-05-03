# Simple Issue Tracker

This is a very simple issue tracker app using Go.

How to run
---
1. Install `Go 1.6 or later` then run `go run main.go`.
2. Open http://localhost:9000/ on your browser

API
---
#### GET ALL ISSUE DATA
http://localhost:9000/
#### GET ISSUE BY ID
http://localhost:9000/issue/{id}
#### CREATE NEW ISSUE
http://localhost:9000/newIssue/
#### UPDATE ISSUE BY ID
http://localhost:9000/change/{id}
#### DELETE ISSUE BY ID
http://localhost:9000/delete/{id}