const fs = require("fs");
const stripe = require("stripe")(
  fs.readFileSync("./config/stripe/key", "utf8").replace(/(\r\n|\n|\r)/gm, "")
);
const express = require("express");

let app = express();

app.get("/api/subscribe/success", async (req, res) => {
  const sessionId = req.query.session_id;

  const session = await stripe.checkout.sessions.retrieve(sessionId);

  if (session) {
    res.redirect("http://localhost:3001/frontend/success");
  } else return res.status(403).send("Invalid checkout session");
});

app.get("/api/subscribe/validate", async (req, res) => {
  const customerId = req.query.customer_id;

  const subscriptions = await stripe.customers.retrieve(customerId, {
    expand: ["subscriptions"],
  });

  let sorted = subscriptions.subscriptions.data.sort(
    (a, b) => a.created - b.created
  );
  sorted = sorted.filter((sub) => sub.status === "active")

  if (!sorted.length)
    return res.status(200).send({ active: false })

  res.status(200).send({ active: true });
});

app.get("/api/subscribe/new/:price", async (req, res) => {
  const { price } = req.params;

  const product = await stripe.products.search({
    query: 'name~"FerretCode Hosting"',
  });

  const prices = await stripe.prices.list({
    product: product.data[0].id,
  });

  const stripePrice = prices.data.find((p) => {
    return p.unit_amount === price * 100; 
  }); 

  const session = await stripe.checkout.sessions.create({
    mode: "subscription",
    line_items: [
      {
        price: stripePrice.id,
        quantity: 1,
      },
    ],
    success_url:
      "http://localhost:3001/api/subscribe/success?session_id={CHECKOUT_SESSION_ID}",
    cancel_url: "http://localhost:3001/frontend/canceled",
  });

  res.redirect(session.url);
});

app.listen(3000);
