import os

PATHS=["github.com"]
GOPATH=os.environ["GOPATH"].split(':;')[0]

for path in PATHS:
	abs_path = os.path.abspath("./" + path)
	for parent in os.scandir("./" + path):
		if not parent.is_dir(): continue
		os.makedirs("{}/src/{}/{}".format(GOPATH, path, parent.name), exist_ok=True)
		for sub in os.scandir(parent.path):
			if not sub.is_dir(): continue
			os.symlink(os.path.abspath(sub.path), "{}/src/{}/{}/{}".format(GOPATH, path, parent.name, sub.name), target_is_directory=True)
