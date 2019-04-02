#!/bin/bash
set -e
set -o pipefail

SCRIPT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/$(basename "${BASH_SOURCE[0]}")"
REPO_URL="${REPO_URL:-r.j3ss.co}"
JOBS=${JOBS:-1}

ERRORS="$(pwd)/errors"


# Set the state directory.
if [[ -z "$STATE_DIR" ]]; then
	STATE_DIR="$(mktemp -d)"
fi

build_and_push(){
	base=$1
	suite=$2
	build_dir=$3

	echo "Building ${REPO_URL}/${base}:${suite} for context ${build_dir}"
	img build -state "$STATE_DIR" -t ${REPO_URL}/${base}:${suite} ${build_dir}

	echo "[In Container] Building ${REPO_URL}/${base}:${suite} for context ${build_dir}"
	# Do the same but in a docker container.
    name=${REPO_URL}-${base}-${suite}
    set +e
	docker run --name $name --volume $(pwd):/home/user/src:ro --workdir /home/user/src --privileged r.j3ss.co/img build --no-console -t ${REPO_URL}/${base}:${suite} ${build_dir}  > /dev/null 2>&1
    status=$?
    set -e
    if [[ $status != 0 ]]; then
        docker logs $name
        docker rm -f $name
        exit $status
    fi
    docker rm -f $name

	# on successful build, push the image
	echo "                       ---                                   "
	echo "Successfully built ${base}:${suite} with context ${build_dir}"
	echo "                       ---                                   "
}

dofile() {
	f=$1
	image=${f%Dockerfile}
	base=${image%%\/*}
	build_dir=$(dirname $f)
	suite=${build_dir##*\/}

	if [[ -z "$suite" ]] || [[ "$suite" == "$base" ]]; then
		suite=latest
	fi

	{
		$SCRIPT build_and_push "${base}" "${suite}" "${build_dir}"
	} || {
		# add to errors
		echo "${base}:${suite}" >> $ERRORS
	}
	echo
	echo
}

main(){
	tmpd=$(mktemp -d)

	# clone the repo
	git clone --depth 1 https://github.com/jessfraz/dockerfiles.git "$tmpd"
	cd "$tmpd"

	# find the dockerfiles
	IFS=$'\n'
	files=( $(find -L . -iname '*Dockerfile' | sed 's|./||' | sort -R --random-source=/dev/urandom | head -n 1) )
	unset IFS

	# build all dockerfiles
	echo "Running in parallel with ${JOBS} jobs."
	parallel --tag --verbose --ungroup -j"${JOBS}" $SCRIPT dofile "{1}" ::: "${files[@]}"

	# cleanup
	rm -rf "$tmpd"

	if [[ ! -f $ERRORS ]]; then
		echo "No errors, hooray!"
	else
		echo "[ERROR] Some images did not build correctly, see below." >&2
		echo "These images failed: $(cat $ERRORS)" >&2
		exit 1
	fi
}

run(){
	args=$@
	f=$1

	if [[ "$f" == "" ]]; then
		main $args
	else
		$args
	fi
}

run $@
