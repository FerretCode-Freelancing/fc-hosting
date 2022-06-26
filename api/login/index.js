const { MyCatLikesFirebaseServer } = require("my-cat-likes-firebase");

const express = require("express");
const jwt = require("jsonwebtoken");
//const bcrypt = require("bcrypt");
const fs = require("fs");

const firebase = new MyCatLikesFirebaseServer({
  firebaseCredentialsPath: "./config/firebase/FIREBASE",
  loggingEnabled: true,
});
firebase.initialize();

function readSecret(path) {
  return Buffer.from(fs.readFileSync(path, "utf8"), "base64").toString("utf8");
}

const app = express();
app.set('trust proxy', 1);
app.use(
  require("express-session")({
    secret: readSecret("./config/session/SESSION"),
    resave: false,
    saveUninitialized: true,
    cookie: { secure: true },
  })
);

app.get("/auth/github", (req, res) => {
  res.redirect(
    `https://github.com/login/oauth/authorize?client_id=${readSecret(
      "./config/gh/id"
    )}&scope=public_repo,`
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
  ).catch(err => {
    console.log(err);
    res.send(500, "Error authenticating")
  });

  const json = await response.json();
  
  if(!json.access_code) {
    res.send(403, "Unauthorized");
    req.session.access_token = null;
  }
  
  res.send(200, "Authenticated");
});

app.listen(3000);
