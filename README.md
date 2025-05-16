# bampro

**Biofoundry And MARL Paragon Repo Orchestrator**

This project leverages Docker Compose to orchestrate the **Primordia simulation** alongside a custom AI backend for distributed multi-agent reinforcement learning (MARL) experiments.

## 📁 Project Structure

```
bampro/
├── docker-compose.yaml    # Docker Compose orchestration file
├── LICENSE               # Project license
├── primordia/            # Primordia simulation/game engine
├── README.md             # This file
├── simulating/           # Custom MARL backend or agent logic
├── thinking/             # Supporting AI logic, neural networks, or Paragon components
```

- **primordia/**: Hosts the simulation engine for running experiments.
- **simulating/**: Contains your custom backend or multi-agent system logic.
- **thinking/**: Includes additional AI components, such as neural networks or MARL algorithms.

## 🚀 Quick Start

To get started, run the following commands in the project root:

```sh
# Stop any running containers
docker-compose down

# Rebuild all images without cache
docker-compose build --no-cache

# Start all services and force recreate containers
docker-compose up --force-recreate
```

## 🖧 Communication

- Services communicate over the Docker Compose network.
- Use **service names** as hostnames (e.g., `primordia:14000`) for connections between the backend and simulation.

## 📝 Useful Commands

- **View logs for a specific service:**

  ```sh
  docker-compose logs primordia
  docker-compose logs simulating
  ```

  Replace with the actual service names defined in `docker-compose.yaml`.

- **Stop all containers:**

  ```sh
  docker-compose down
  ```

- **Clean up stopped containers, networks, and images (optional):**

  ```sh
  docker system prune -af
  ```

## ⚡️ Tips

- Modify ports or add services by editing `docker-compose.yaml`.
- Always use **Docker service names** instead of `localhost` for inter-container communication.
- To run commands without `sudo`, ensure your user is added to the `docker` group:

  ```sh
  sudo usermod -aG docker $USER
  ```
