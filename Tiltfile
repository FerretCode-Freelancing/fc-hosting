load('ext://restart_process', 'docker_build_with_restart')

docker_build_with_restart('sthanguy/fc-auth',
							context='./services/auth',
							entrypoint='node .',
							dockerfile='./services/auth/Dockerfile',
							extra_tag='latest',
							live_update=[
								sync('./services/auth', '/home/nonroot/route'),
							]
)

docker_build_with_restart('sthanguy/fc-gateway',
							context='./services/gateway',
							entrypoint='go run main.go',
							dockerfile='./services/gateway/Dockerfile',
							extra_tag='latest',
							live_update=[
								sync('./services/gateway', '/usr/gateway'),
							]
)

docker_build_with_restart('sthanguy/fc-subscribe',
							context='./services/subscribe',
							entrypoint='node .',
							dockerfile='./services/subscribe/Dockerfile',
							extra_tag='latest',
							live_update=[
								sync('./services/subscribe', '/home/nonroot/route'),
							]
)

docker_build_with_restart('sthanguy/fc-upload',
							context='./services/upload',
							entrypoint='go run main.go',
							dockerfile='./services/upload/Dockerfile',
							extra_tag='latest',
							live_update=[
								sync('./services/upload', '/home/nonroot/route'),
								run('cd /home/nonroot/route && go build -v -o /home/nonroot/route'),
								run('route')
							]
)

docker_build_with_restart('sthanguy/fc-projects',
							context='./services/projects',
							entrypoint='go run main.go',
							dockerfile='./services/projects/Dockerfile',
							extra_tag='latest',
							ignore=['.kubemqctl.yaml'],
							live_update=[
								sync('./services/projects', '/home/nonroot/route'),
								run('cd /home/nonroot/route && go build -v -o /home/nonroot/route'),
								run('route')
							]
)

docker_build_with_restart('sthanguy/fc-provision',
							context='../fc-provision',
							dockerfile='../fc-provision/Dockerfile',
							entrypoint='builder',
							extra_tag='sthanguy/fc-provision:latest',
							live_update=[
								sync('../fc-provision', '/usr/src/provision'),
								run('cd /usr/src/provision && go build -v -o /usr/local/bin/builder'),
							]
)

docker_build_with_restart('sthanguy/fc-deploy',
							context='./services/deploy',
							dockerfile='./services/deploy/Dockerfile',
							entrypoint='go run main.go',
							extra_tag='latest',
							ignore=['.kubemqctl.yaml'],
							live_update=[
								sync('./services/deploy', '/home/nonroot/route'),
								run('cd /home/nonroot/route && go build -v -o /home/nonroot/route'),
								run('route')
							]
)

# deployments
k8s_yaml(['kubernetes/deployments/auth.yml', 'kubernetes/deployments/gateway.yml', 'kubernetes/deployments/subscribe.yml', 'kubernetes/deployments/upload.yml', 'kubernetes/deployments/cache.yml', 'kubernetes/deployments/registry.yml', 'kubernetes/deployments/provision.yml', 'kubernetes/deployments/projects.yml', 'kubernetes/deployments/deploy.yml'])
# services
k8s_yaml(['kubernetes/services/auth.yml', 'kubernetes/services/gateway.yml', 'kubernetes/services/subscribe.yml', 'kubernetes/services/upload.yml', 'kubernetes/services/cache.yml', 'kubernetes/services/registry.yml', 'kubernetes/services/provision.yml', 'kubernetes/services/projects.yml', 'kubernetes/services/deploy.yml'])
# pvcs
k8s_yaml(['kubernetes/pvc/registry.yml'])
# ingress
k8s_yaml(['kubernetes/ingress/gateway.yml'])
# roles
k8s_yaml(['kubernetes/roles/deploy.yml'])
