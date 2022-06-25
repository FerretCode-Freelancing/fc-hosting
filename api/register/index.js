const { MyCatLikesFirebaseServer } = require("my-cat-likes-firebase");

const express = require("express");
const jwt = require("jsonwebtoken");
//const bcrypt = require("bcrypt");
const passport = require("passport");
const GithubStrategy = require("passport-github2");
const fs = require("fs");

const firebase = new MyCatLikesFirebaseServer({
  firebaseCredentialsPath: "./config/firebase/FIREBASE",
  loggingEnabled: true,
});
firebase.initialize();

function readSecret(path) {
  return Buffer.from(
    fs.readFileSync(path, "utf8"),
    "base64"
  ).toString("utf8")
}

passport.use(
  new GithubStrategy.Strategy(
    {
      clientId: readSecret("./config/gh/id"),
      clientSecret: readSecret("./config/gh/secret"),
      callbackURL: "http://127.0.0.1:3000/auth/github/callback",
    },
    (accessToken, refreshToken, profile, done) => {
      firebase
        .getDoc(`users/${profile.id}`)
        .then((user) => {
          if (!user)
            firebase
              .createDoc({}, `users/${profile.id}`)
              .then(() => done(null, {}))
              .catch((err) => done(err, null));

          return done(null, user);
        })
        .catch((err) => done(err, null));
    }
  )
);

passport.serializeUser((user, done) => {
  done(null, user);
});

passport.deserializeUser((user, done) => {
  done(null, user);
});

const app = express();
app.use(
  require("express-session")({
    secret: readSecret("./config/session/SESSION"),
    resave: true,
    saveUninitialized: true,
  })
);

app.get("/auth/github", passport.authenticate("github"));

app.get(
  "/auth/github/callback",
  passport.authenticate("github", { failureRedirect: "/auth/failure" }),
  (req, res) => {
    res.send(JSON.stringify(req.headers, null, 2));
  }
);

app.listen(3000);
