import os


def check_yml(file):
    if(file.name.find(".yml") == -1):
        return False
    return True


for dir in os.scandir('../kubernetes'):
    for config in filter(check_yml, os.scandir(dir.path)):
        os.system(f'sudo kubectl apply -f {config.path}')
