const fs = require("fs");
const { exec } = require("child_process");
 
for(const dir in fs.readdirSync("../api")) {
  exec(
    `docker build -t sthanguy/fc-${dir}:latest .`, 
    (err, output) => {
      if(err) console.error(err);
      
      console.log(output);
    }
  )
}

for(const dir in fs.readdirSync("../api")) {
  exec(
    `docker publish sthanguy/fc-${dir}:latest .`,
    (err, output) => {
      if(err) console.error(err);
      
      console.log(output);
    }
  )
}