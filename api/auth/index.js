const { MyCatLikesFirebaseServer } = require("my-cat-likes-firebase");

const express = require("express");
const session = require("express-session");
const bcrypt = require("bcrypt");
const fs = require("fs");

const RedisStore = require("connect-redis")(session);
const { createClient } = require("redis");

const redisClient = createClient({
  legacyMode: true,
  url: `redis://${process.env.FC_SESSION_STORAGE_SERVICE_HOST}:${process.env.FC_SESSION_STORAGE_SERVICE_PORT}`,
})
redisClient.connect().catch(console.error)

const firebase = new MyCatLikesFirebaseServer({
  firebaseCredentialsPath: "./config/firebase/FIREBASE",
  loggingEnabled: true,
});

function readSecret(path) {
  return fs.readFileSync(path, "utf-8").replace(/(\r\n|\n|\r)/gm, "");
}

function catchError(err, res) {
  console.log(err);

  if (res.headersSent) return;

  return res.status(500).send("Internal server error. Please try again later.");
}

const app = express();
app.set("trust proxy", 1);
app.use(
  session({
    name: "fc-hosting",
    secret: readSecret("./config/session/secret"),
    resave: false,
    store: new RedisStore({ client: redisClient }),
    saveUninitialized: false,
    cookie: { secure: false }, //TODO: set to true when https is enabled
  })
);

const sessionString = {
	start: 4,
	end: 36,
	prefix: "sess:"
}; 

app.get("/auth/github/user", async (req, res) => {
	const id = req.session.id;

	const sess = await redisClient.hGet(`${sessionString.prefix}${
		id.slice(sessionString.start, sessionString.end)
	}`).catch((err) => res.status(500).send(err));

	if(sess && !res.headersSent)
		return res.status(403).send("Failed to validate auth.");

	const token = req.session.access_token;	

	const user = await fetch("https://api.github.com/user", {
		headers: {
			Accept: "application/json",
			Authorization: `token ${token}`,
		},
	}).catch((err) => catchError(err, res));

	const userJson = await user.json();

	res.status(200).send({ id: userJson.id });
});

app.get("/auth/github", (_, res) => {
  const url = `https://github.com/login/oauth/authorize?client_id=${readSecret(
    "./config/gh/id"
  )}&scope=public_repo,read:user,user:email&redirect_uri=http://localhost:3001/auth/github/callback`;

  res.redirect(url);
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

  if (!loginJson.access_token) {
    req.session.access_token = null;
    return res
      .status(403)
      .send({
        error: "There was an issue authenticating you! Please try again later.",
      });
  }

  req.session.access_token = loginJson.access_token;

  req.session.save();

  const user = await fetch("https://api.github.com/user", {
    headers: {
      Accept: "application/json",
      Authorization: `token ${loginJson.access_token}`,
    },
  }).catch((err) => catchError(err, res));

  const email = await fetch("https://api.github.com/user/emails", {
    headers: {
      Accept: "application/json",
      Authorization: `token ${loginJson.access_token}`,
    },
  }).catch((err) => catchError(err, res));

  const userJson = await user.json();
  const emailJson = await email.json();

  const salt = await bcrypt.genSalt(10).catch((err) => catchError(err, res));
  const hash = await bcrypt
    .hash(emailJson.find((e) => e.primary === true).email, salt)
    .catch((err) => catchError(err, res));

  firebase.findOrCreateDoc()
    .then(() => { });

  firebase
    .findOrCreateDoc(
      {
        projects: [],
        runningProjects: [],
        email: hash,
      },
      `users/${userJson.id}`
    )
    .then(async () => {
      const response = await fetch("http://localhost:5000/", {
        method: "POST",
        body: JSON.stringify({
          message: `User ${userJson.id} has logged in`,
          webhook: true,
        }),
        headers: {
          "Content-Type": "application/json",
        },
      }).catch((err) => catchError(err, res));

      const json = await response.text();

      console.log(json);
    })
    .catch((err) => catchError(err, res));

  if (!res.headersSent)
    res
      .status(200)
      .send({ message: "You have been successfully authenticated." });
});

app.listen(3000);
