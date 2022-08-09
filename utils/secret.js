const fs = require("fs");
const readline = require("node:readline");

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
});

function question(question) {
  return new Promise((resolve) => {
    rl.question(`${question}\n`, (answer) => {
      resolve(answer);
    })
  })
}

function parseData(data) {
  try {
    JSON.parse(data);
    return "stringData:"
  } catch(err) {
    return "data:"
  }
}

(async () => {
  const fileName = await question("What is the name of the file?");
  const name = await question("What is the name of the secret?");
  const data = await question("What is the data in the secret?");
 
  const template = `
    apiVersion: v1
    kind: Secret
    metadata:
      name: ${name} 
    type: Opaque
    ${parseData(data)}
      ${data}    
  `;

  const confirmation = await question("Are you sure you want to create this secret? y/n");   
  
  confirmation ? 
    fs.writeFileSync(`../kubernetes/secrets/${fileName}`, template) : 
    process.exit(0);

  console.log("The secret was created.");
  
  process.exit(0);
})();
