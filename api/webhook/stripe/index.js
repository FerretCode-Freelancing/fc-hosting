const fs = require("fs");
const stripe = require("stripe")(
  Buffer.from(
    fs.readFileSync("../config/stripe/KEY"),
    "base64"
  ).toString("utf8")
);

module.exports.StripeWebhook = class {
  constructor(req, res, firebase) {
    this.req = req;
    this.res = res;
    this.firebase = firebase;
    
    let data;
    let eventType;
    
    const webhookSecret = Buffer.from(
      fs.readFileSync("../config/stripe/WEBHOOK_SECRET", "utf8"),
      "base64"
    ).toString("utf8");
    
    let event;
    let signature = this.req.headers["stripe-signature"];
    
    try {
      event = stripe.webhooks.constructEvent(
        this.req.body,
        signature,
        webhookSecret
      );
    } catch (err) {
      console.log("Webhook verification failed");
      this.res.send(400, "Failed to verify webhook");
    }
    
    switch(eventType) {
      case "checkout.session.completed":
      break;
      
      case "invoice.paid":
      break;
    
      case "invoice.payment_failed":
      break;
    
      default:
        res.send(403, "Payment was not recieved.");
    }
  }
}