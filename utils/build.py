import os

# build
print("BUILDING")

for dir in os.scandir('../services'):
    print(dir.path)

    image_name = f'sthanguy/fc-{dir.path[12:]}'
    build_cmd = f'sudo docker build -t {image_name} {dir.path}'
    os.system(build_cmd)

# publish
print("PUBLISHING")

to_publish = []

for dir in os.scandir('../services'):
    pub = input(f'would you like to publish {dir.path}?\n')

    to_publish.append(dir.path) if pub == 'y' else ''

for dir in to_publish:
    print(dir)

    image_name = f'sthanguy/fc-{dir[12:]}'
    publish_cmd = f'sudo docker push {image_name}'
    os.system(publish_cmd)

# restart
for dir in to_publish:
    cmd = f'sudo kubectl rollout restart deployment fc-{dir[12:]}'
    os.system(cmd)

os.system("sudo kubectl get pods")
