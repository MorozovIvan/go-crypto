.PHONY: start-backend stop-backend start-frontend stop-frontend restart clean check-ports kill-all check-backend-port check-frontend-port

# Kill all related processes
kill-all:
	@echo "Killing all related processes..."
	@pkill -f "go run main.go" || true
	@pkill -f "npm run dev" || true
	@lsof -ti:8080,8081,5173 | xargs kill -9 2>/dev/null || true
	@rm -f .backend.pid .frontend.pid
	@sleep 2

# Check if backend port is available
check-backend-port:
	@echo "Checking backend port..."
	@if lsof -ti:8080 > /dev/null; then \
		echo "Port 8080 is in use. Cleaning up..."; \
		make stop-backend; \
	fi
	@sleep 2

# Check if frontend port is available
check-frontend-port:
	@echo "Checking frontend port..."
	@if lsof -ti:5173 > /dev/null; then \
		echo "Port 5173 is in use. Cleaning up..."; \
		make stop-frontend; \
	fi
	@sleep 2

# Start the backend server
start-backend: check-backend-port
	@echo "Starting backend server..."
	@go run main.go & echo $$! > .backend.pid
	@echo "Waiting for backend server to start..."
	@for i in 1 2 3 4 5; do \
		if curl -s http://localhost:8080/api/telegram/test > /dev/null; then \
			echo "Backend server is running on port 8080"; \
			break; \
		fi; \
		if [ $$i -eq 5 ]; then \
			echo "Failed to start backend server"; \
			make stop-backend; \
			exit 1; \
		fi; \
		sleep 2; \
	done

# Stop the backend server
stop-backend:
	@echo "Stopping backend server..."
	@if [ -f .backend.pid ]; then \
		kill -9 $$(cat .backend.pid) 2>/dev/null || true; \
		rm .backend.pid; \
	fi
	@lsof -ti:8080,8081 | xargs kill -9 2>/dev/null || true
	@sleep 2

# Start the frontend server
start-frontend: check-frontend-port
	@echo "Starting frontend server..."
	@cd frontend && npm run dev & echo $$! > .frontend.pid
	@echo "Waiting for frontend server to start..."
	@for i in 1 2 3 4 5; do \
		if curl -s http://localhost:5173 > /dev/null; then \
			echo "Frontend server is running on port 5173"; \
			break; \
		fi; \
		if [ $$i -eq 5 ]; then \
			echo "Failed to start frontend server"; \
			make stop-frontend; \
			exit 1; \
		fi; \
		sleep 2; \
	done

# Stop the frontend server
stop-frontend:
	@echo "Stopping frontend server..."
	@if [ -f .frontend.pid ]; then \
		kill -9 $$(cat .frontend.pid) 2>/dev/null || true; \
		rm .frontend.pid; \
	fi
	@lsof -ti:5173 | xargs kill -9 2>/dev/null || true
	@sleep 2

# Restart both servers
restart: kill-all
	@echo "Restarting servers..."
	@make start-backend
	@make start-frontend

# Clean up all processes
clean: kill-all

# Start both servers
start: kill-all
	@echo "Starting servers..."
	@make start-backend
	@make start-frontend

# Stop both servers
stop: kill-all

# Help command
help:
	@echo "Available commands:"
	@echo "  make start-backend    - Start the backend server"
	@echo "  make stop-backend     - Stop the backend server"
	@echo "  make start-frontend   - Start the frontend server"
	@echo "  make stop-frontend    - Stop the frontend server"
	@echo "  make restart         - Restart both servers"
	@echo "  make clean           - Clean up all server processes"
	@echo "  make start           - Start both servers"
	@echo "  make stop            - Stop both servers"
	@echo "  make kill-all        - Force kill all related processes"
	@echo "  make help            - Show this help message" 