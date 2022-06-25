const { MyCatLikesFirebaseServer } = require("my-cat-likes-firebase");

const express = require("express");
const jwt = require("jsonwebtoken");
const bcrypt = require("bcrypt");
const passport = require("passport");
const GithubStrategy = require("passport-github2");
const fs = require("fs");

const firebase = new MyCatLikesFirebaseServer({
  firebaseCredentialsPath: "./config/FIREBASE",
  loggingEnabled: true,
});

passport.use(
  new GithubStrategy.Strategy({
    clientId: fs.readFileSync("./config/gh/id"),
    clientSecret: fs.readFileSync("./config/gh/secret"),
    callbackURL: "http://127.0.0.1:3000/dashboard",
  }),
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
);

const app = express();

app.get(
  "/auth/github",
  passport.authenticate("github", { scope: ["user:repo"] })
);

app.get(
  "/auth/github/callback",
  passport.authenticate("github", { failureRedirect: "/auth/failure" }),
  (req, res) => {
    res.redirect(`/dashboard?acccess_token=${req.body.access_token}&refresh_token=${req.body.refresh_token}`);
  }
);
