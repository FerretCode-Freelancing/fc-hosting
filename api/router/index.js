const { millie } = require("millie");
const routes = require("./routes.json");
const url = require("node:url");

const app = new millie(3000);
app.initialize();

app.server.on("request", (req, res) => new Request(req, res).proxy());

class Request {
  constructor(req, res) {
    this.proxy = () => this.request(req, res); 
    
    this.request = async (req, res) => {
      const { endpoint } = getEndpoint(req);
      const query = url.parse(req.url, true).query;  
      
      if(!endpoint)
        return res.respond(404, "Endpoint not found.");

      const data = await fetch(
        `http://${process.env[`FC_${endpoint}_SERVICE_HOST`]}:${
          process.env[`FC_${endpoint}_SERVICE_PORT`]
        }${req.url}${Object.keys(query).length > 0 ? "?" : ""}${new URLSearchParams({
          ...query,
        }).toString()}`,
        { body: req.body, method: req.method, redirect: 'follow' }
      ).catch((err) => handleError(err, res));  
      
      const json = await data.json();
      
      if(json.url)
        return res.writeHead(302, {
          'Location': json.url
        }).end(); 

      res.respond(data.status, json);
    };
  }
}

function handleError(err, res) {
  console.error(err);

  return res.respond(500, "Internal server error. Please try again later.");
}

function getEndpoint(req) {
  const route = req.url.slice(1);
  const base = routes.routes[route.slice(0, route.indexOf("/"))];

  if (!Array.isArray(base)) return { endpoint: base };

  const service = req.url
    .slice(1)
    .slice(req.url.indexOf("/"), req.url.lastIndexOf("/"))
    .toUpperCase();

  return { endpoint: service };
}
