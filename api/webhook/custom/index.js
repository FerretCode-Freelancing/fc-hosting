const fs = require("fs");
const crypto = require("crypto");

module.exports.CustomWebhook = class {
  constructor(req, res, body, url) {
    this.req = req;
    this.res = res;
    this.body = body;
    this.url = url;

    const token = this.req.secret;
    const actualToken = Buffer.from(
      fs.readFileSync("../config/custom/token", "utf8"),
      "base64"
    ).toString("utf8");

    if (crypto.timingSafeEqual(token, actualToken)) {
      fetch(url, {
        method: "POST",
        body: this.body,
      })
      .then(() => res.send("Webhook sent"))
      .catch((err) => {
        console.log(err);
        res.send(500, "Internal server error");
      });
    } else { 
      res.send(403, "Invalid webhook token");
    }
  }
}