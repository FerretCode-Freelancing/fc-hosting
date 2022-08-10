const fs = require("fs");
const stripe = require("stripe")(
  Buffer.from(
    fs.readFileSync("./config/stripe/KEY", "utf8"),
    "base64"
  ).toString("utf8")
);
const express = require("express");

let app = express();

app.post("/api/subscribe", async (req, res) => {
  const { price } = req.body;

  const session = await stripe.checkout.sessions.create({
    mode: "subscription",
    line_items: [
      {
        price,
        quantity: 1,
      },
    ],
    success_url: "http://127.0.0.1:3000/success?session_id={CHECKOUT_SESSION_ID}",
    cancel_url: "http://127.0.0.1:3000/canceled"
  });
  
  res.redirect(303, session.url);
});

app.listen(3000);
