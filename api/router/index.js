const express = require("express");

const app = express();

app.get("/api/*", async (req, res) => {
  const { endpoint, baseEndpoint } = getEndpoint(req)

  const data = await fetch(
    `${baseEndpoint}_SERVICE_HOST:${baseEndpoint}_SERVICE_PORT/${endpoint.toLowerCase()}?${new URLSearchParams({ ...req.query })}`,
    { headers: req.headers, body: req.body }
  ).catch(err => handleError(err, res));
  
  const json = data.json().catch(err => handleError(err, res));
  
  res.status(data.status).send(json);
});

app.post("/api/*", async (req, res) => {
  const { endpoint, baseEndpoint } = getEndpoint(req);

  const data = await fetch(
    `${baseEndpoint}_SERVICE_HOST:${baseEndpoint}_SERVICE_PORT/${endpoint.toLowerCase()}?${new URLSearchParams({ ...req.query })}`,
    { headers: req.headers, body: req.body }
  ).catch(err => handleError(err, res));
  
  const json = data.json().catch(err => handleError(err, res));
  
  res.status(data.status).send(json);
});

function handleError(err, res) {
  console.error(err);
  
  return res.status(500).send("Internal server error. Please try again later.");
}

function getEndpoint(req) {
  const endpoint = req.originalUrl
    .slice(1, req.originalUrl.indexOf("/"))
    .toUpperCase();
  const baseEndpoint = req.baseUrl().slice(1);
  
  return { endpoint, baseEndpoint };
}  

app.listen(3000);
console.log("Listening on 3000");
