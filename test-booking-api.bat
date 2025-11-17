@echo off
echo Testing Event Detail and Booking API...
echo.

echo 1. Testing GET /event/1 (should return event details)
curl -X GET http://localhost:8080/event/1
echo.
echo.

echo 2. Testing POST /bookings (create a booking)
curl -X POST http://localhost:8080/bookings ^
  -H "Content-Type: application/json" ^
  -d "{\"event_id\":1,\"name\":\"John Doe\",\"email\":\"john@example.com\",\"phone\":\"+1234567890\",\"tickets\":2,\"notes\":\"Looking forward to it!\"}"
echo.
echo.

echo Test complete!
pause
