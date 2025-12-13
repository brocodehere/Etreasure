# Ethnic Treasures – Razorpay Test Integration

## Payment Integration (Test Mode)

This project uses **Razorpay** as the only payment gateway (test mode only).

### Environment Variables

#### Backend (`/backend/.env` or `/backend/.env.sample`):
```
RAZORPAY_KEY_ID=rzp_test_xxxxxxxxxxxx
RAZORPAY_KEY_SECRET=xxxxxxxxxxxxxxxx
```

#### Frontend (`/web/.env` or `/web/.env.sample`):
```
VITE_RAZORPAY_KEY_ID=rzp_test_xxxxxxxxxxxx
```

### Backend Endpoints

- `POST /api/orders/create-payment` — Creates a Razorpay order for the cart, returns Razorpay order ID and key.
- `POST /api/orders/verify-payment` — Verifies Razorpay payment signature and marks order as paid.

### How to Run Locally

#### Backend
```
cd backend
# Set up .env with Razorpay test keys and database config
# Run migrations if needed
# Start API server:
go run ./cmd/api
```

#### Frontend
```
cd web
# Set up .env with VITE_RAZORPAY_KEY_ID
npm install
npm run dev
```

### How to Test Razorpay Payment (Test Mode)
1. Add products to cart and proceed to checkout.
2. Fill shipping info. On the Payment step, click **Pay Securely with Razorpay**.
3. Use Razorpay test cards (e.g. `4111 1111 1111 1111`, any future date, any CVV).
4. On success, order is marked paid and you are redirected to the confirmation page.
5. If payment fails or is closed, order remains pending and you can retry.

> **Note:** Only Razorpay is available. All other payment options have been removed.

---

## Developer Notes
- Razorpay is in **test mode**. Do not use real cards.
- No payment secret is ever exposed on frontend.
- If you change payment logic, update this README.
