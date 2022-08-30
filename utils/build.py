import os

# build
print("BUILDING")

for dir in os.scandir('../api'):
    print(dir.path)

    image_name = f'sthanguy/fc-{dir.path[7:]}'
    build_cmd = f'sudo docker build -t {image_name} {dir.path}'
    os.system(build_cmd)

# publish
print("PUBLISHING")

to_publish = []

for dir in os.scandir('../api'):
    pub = input(f'would you like to publish {dir.path}?\n')

    to_publish.append(dir.path) if pub == 'y' else ''

for dir in to_publish:
    print(dir)

    image_name = f'sthanguy/fc-{dir[7:]}'
    publish_cmd = f'sudo docker push {image_name}'

    os.system(publish_cmd)

print("RESTARTING")

for dir in to_publish:
    deployment_name = f'fc-{dir[7:]}'

    os.system(f'kubectl rollout restart deployment {deployment_name}')

print("PODS")

os.system('kubectl get pods')
