# Campuswire reputation

This is go script to fetch Campuswire reputation and store it to a google sheet.

Three environment variables are needed

- `SHEETID`: the google sheet database
- `SHEETSECRET`: a base64 encoded service account json
- `CWTOKEN`: the campuswire access token, it is obtained by inspecting the http requests made by the browser

