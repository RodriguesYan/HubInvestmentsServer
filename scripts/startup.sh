#!/bin/bash
# Kill any process using ports
lsof -ti:50051 | xargs kill -9 2>/dev/null
lsof -ti:8080 | xargs kill -9 2>/dev/null
sleep 2

# Start app
cd /Users/yanrodrigues/Documents/HubInvestmentsProject/HubInvestmentsServer
/tmp/hub_test > /tmp/app.log 2>&1 &

echo "App started. PID: $!"
sleep 5
ps aux | grep hub_test | grep -v grep && echo "✅ App is running"
