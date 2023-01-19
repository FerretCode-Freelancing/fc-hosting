const fs = require("fs");
const stripe = require("stripe")(
  fs.readFileSync("./config/stripe/key", "utf8").replace(/(\r\n|\n|\r)/gm, "")
);
const firebase = require("my-cat-likes-firebase");
const express = require("express");

const firestore = new firebase.MyCatLikesFirebaseServer({
  firebaseCredentialsPath: "./config/firebase/FIREBASE",
  loggingEnabled: false,
});

let app = express();

app.get("/api/subscribe/success", async (req, res) => {
  const sessionId = req.query.session_id;
  const projectId = req.query.project_id;

  const session = await stripe.checkout.sessions.retrieve(sessionId, {
    expand: ["subscription"],
  });

  if (session) {
    if (!session.subscription)
      return res.status(403).send("The checkout failed.");

    try {
      const owner = await fetch(
        `http://${process.env.FC_AUTH_SERVICE_HOST}:${process.env.FC_AUTH_SERVICE_PORT}/auth/user`
      );

      const json = await owner.json();

      let limit;
      switch (subscription.items.data[0].unit_amount) {
        case 100:
          limit = 100;
          break;
        case 200:
          limit = 250;
          break;
        case 500:
          limit = 500;
          break;
      }

      firestore.updateDoc(
        {
          projects: {
            [projectId]: {
              subscription_id: session.subscription,
              ram_limit: limit,
            },
          },
        },
        `users/${json.owner_id}`
      );

      res.status(200).send("The checkout succeeded.");
    } catch (err) {
      res.status(500).send("There was an error validating the subscription.");
    }
  } else return res.status(403).send("Invalid checkout session");
});

app.get("/api/subscribe/validate", async (req, res) => {
  const customerId = req.query.customer_id;
  const projectId = req.query.project_id;

  try {
    const subscriptions = await stripe.customers.retrieve(customerId, {
      expand: ["subscriptions"],
    });

    const owner = await fetch(
      `http://${process.env.FC_AUTH_SERVICE_HOST}:${process.env.FC_AUTH_SERVICE_PORT}/auth/user`
    );

    const json = await owner.json();

    const user = await firestore.getDoc(`users/${json.owner_id}`);
    const project = user.projects[projectId];

    const active = subscriptions.subscriptions.data.some(
      (subscription) => subscription.id === project.subscriptionId
    );

    if (active) return res.status(200).send({ active });

    res.status(200).send({ active });
  } catch (err) {
    res.status(500).send("There was an error validating the subscription.");
  }
});

app.get("/api/subscribe/new/:price", async (req, res) => {
  const { price } = req.params;
  const { projectId } = req.query;

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
    success_url: `http://localhost:1337/api/subscribe/success?session_id={CHECKOUT_SESSION_ID}&project_id=${projectId}`,
    cancel_url: "http://localhost:1337/frontend/canceled",
  });

  res.redirect(session.url);
});

app.listen(3000);
