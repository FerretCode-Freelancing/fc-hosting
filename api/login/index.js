const { MyCatLikesFirebaseServer } = require("my-cat-likes-firebase");

const express = require("express");
const jwt = require("jsonwebtoken");
const bcrypt = require("bcrypt");
const fs = require("fs");

const firebase = new MyCatLikesFirebaseServer({
  firebaseCredentialsPath: "./config/firebase/FIREBASE",
  loggingEnabled: true,
});

function readSecret(path) {
  return Buffer.from(fs.readFileSync(path, "utf8"), "base64").toString("utf8");
}

function catchError(err, res) {
  console.log(err);
  res.send(500, "Internal server error. Please try again later.");
}

const app = express();
app.set("trust proxy", 1);
app.use(
  require("express-session")({
    secret: readSecret("./config/session/secret"),
    resave: false,
    saveUninitialized: true,
    cookie: { secure: true },
  })
);

app.get("/auth/github", (_, res) => {
  res.redirect(
    `https://github.com/login/oauth/authorize?client_id=${readSecret(
      "./config/gh/id"
    )}&scope=public_repo,read:user,user`
  );
});

app.get("/auth/github/callback", async (req, res) => {
  const code = req.query.code;

  const response = await fetch(
    `https://github.com/login/oauth/access_token?client_id=${readSecret(
      "./config/gh/id"
    )}&client_secret=${readSecret("./config/gh/secret")}&code=${code}`,
    {
      method: "POST",
      headers: {
        Accept: "application/json",
      },
    }
  ).catch((err) => catchError(err, res));

  const loginJson = await response.json();

  if (!loginJson.access_code) {
    res.send(403, "Unauthorized");
    req.session.access_code = null;
  }

  req.session.access_code = loginJson.access_code;

  const user = await fetch("https://api.github.com/user", {
    headers: {
      Accept: "application/json",
      Authorization: `token ${loginJson.access_code}`,
    },
  }).catch((err) => catchError(err, res));

  const userJson = await user.json();
  
  const salt = await bcrypt.genSalt(10).catch((err) => catchError(err, res)); 
  const hash = await bcrypt.hash(userJson.email, salt).catch((err) => catchError(err, res));

  firebase
    .findOrCreateDoc({
      projects: [],
      runningProjects: [],
    }, `users/${hash}`)
    .then((data) => {
      //TODO: send webhook
    })
    .catch((err) => catchError(err, res));

  res.send(200, "Authenticated");
});

app.listen(3000);
