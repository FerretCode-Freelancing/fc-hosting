const { GithubWebhook } = require("./gh/index");
const { StripeWebhook } = require("./stripe/index");
const { CustomWebhook } = require("./custom/index");
const { MyCatLikesFirebaseServer } = require("my-cat-likes-firebase");

const express = require("express");
const fs = require("fs");

const firebase = new MyCatLikesFirebaseServer({
  firebaseCredentialsPath: "./config/firebase/FIREBASE",
  loggingEnabled: true
});
const app = express();

app.post("/api/webhook", (req, res) => {
  if(req.headers["stripe-signature"])
    new StripeWebhook(req, res, firebase);
  else if(!req.headers.custom)
    new GithubWebhook(req, res);
  else
    new CustomWebhook(
      req,
      res,
      req.body, 
      Buffer.from(
        fs.readFileSync("./config/discord/WEBHOOK_URL", "utf8"),
        "base64"
      ).toString("utf8")
    );
});