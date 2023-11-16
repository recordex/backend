.PHONY: deploy-dev
deploy-dev:
	git checkout dev
	git fetch origin
	git reset --hard origin/main
	git merge ${branch}
	git log -n 2
	git push -f origin dev
	git checkout ${branch}

.PHONY: generate_proto
generate_proto:
	sh scripts/gen_proto.sh
