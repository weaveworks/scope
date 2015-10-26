from invoke import task, run
import requests

@task
def update():
    r = requests.get('https://mkcert.org/generate/')
    r.raise_for_status()
    certs = r.content

    with open('certifi.go', 'rb') as f:
        file = f.read()

    file = file.split('`\n')
    assert len(file) == 3
    file[1] = certs

    run("rm certifi.go")

    with open('certifi.go', 'wb') as f:
        f.write('`\n'.join(file))
