.PHONY: run stop restart logs clean build deploy undeploy

# --- Docker Compose (local) ---

run:
	docker compose up --build -d
	@echo "AIBoard running at http://localhost:8080"

stop:
	docker compose down

restart:
	docker compose down
	docker compose up --build -d
	@echo "AIBoard running at http://localhost:8080"

logs:
	docker compose logs -f

clean:
	docker compose down -v

build:
	docker compose build

# --- Kubernetes ---

deploy:
	kubectl apply -k k8s/
	@echo "Waiting for rollout..."
	kubectl -n aiboard rollout status statefulset/postgres --timeout=120s
	kubectl -n aiboard rollout status deployment/aiboard --timeout=120s
	@echo "AIBoard deployed. Run 'make k8s-port-forward' to access locally."

undeploy:
	kubectl delete -k k8s/

k8s-port-forward:
	@echo "AIBoard available at http://localhost:8080"
	kubectl -n aiboard port-forward svc/aiboard 8080:80

k8s-logs:
	kubectl -n aiboard logs -l app=aiboard -f
