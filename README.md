---

# **Mobile Money Payment CLI (Go + Campay API)**

This project is a simple **command-line Go application** that allows users to initiate and track **MTN Mobile Money payments** using the **Campay API**.
It securely loads API credentials from a `.env` file and performs a real transaction request + status polling.

---

## 🚀 Features

* Prompts user for:

  * Mobile Money number
  * Amount
  * Description
* Sends a transaction request to **Campay** (`/collect/` endpoint)
* Polls the transaction status (`/transaction/<reference>/`)
* Displays final status:

  * **SUCCESSFUL**
  * **FAILED**
  * (Loops while **PENDING**)
* Loads API key securely from `.env`
* Uses Resty HTTP client

---

## 📦 Requirements

Ensure you have:

* Go 1.18+
* Internet connection
* Campay API Key
* MTN Mobile Money active

Install dependencies:

```sh
go get github.com/joho/godotenv
go get github.com/go-resty/resty/v2
```

---

## 🔐 Environment Setup

Create a `.env` file in the root directory of the project:

```
API_KEY="your-campay-api-key"
```

⚠️ **Do NOT upload `.env` to GitHub.**
Add it to `.gitignore` if needed.

---

## ▶️ Running the Program

```sh
go run main.go
```

You will be prompted for:

```
Enter mobile money number:
Enter amount:
Enter description:
```

The app then:

1. Sends a payment request
2. Displays your transaction reference
3. Checks transaction status every 3 seconds
4. Prints the final result

---

## 📡 API Endpoints Used

### **1. Initiate payment**

`POST https://campay.net/api/collect/`

Body example:

```json
{
  "amount": "500",
  "from": "677123456",
  "description": "Payment test"
}
```

### **2. Check payment status**

`GET https://campay.net/api/transaction/<reference>/`

Possible responses:

* `PENDING`
* `SUCCESSFUL`
* `FAILED`

---

## 🧱 Project Structure

```
project-folder/
│── main.go
│── go.mod
│── go.sum
└── .env
```

---

## 🛡️ Security Notes

* Do **not** hard-code your API key.
* Always use environment variables (`.env` file).
* If you publish this project, **exclude `.env`**.
* Never commit real API keys to GitHub.

---



