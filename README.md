# swarm-cron 

`swarm-cron` is a service designed for Docker Swarm that allows you to schedule containers to run at specified intervals by simply adding a label to your service definitions. This service is especially useful for executing periodic tasks within your Swarm cluster, such as backups, cleanup jobs, or any custom tasks that need to run on a schedule.

## How It Works

`swarm-cron` monitors your Swarm services for the `cronjob_schedule` label. When it finds a service with this label, it schedules the service to run according to the cron schedule specified in the label's value. The service is designed to work exclusively with Docker Swarm and leverages the Swarm API to manage service tasks.

## Usage

To use `swarm-cron` in your Docker Swarm cluster, follow these steps:

1. **Deploy swarm-cron Service:**

   First, you need to deploy the `swarm-cron` service to your Docker Swarm cluster. It's recommended to deploy this service on a manager node with a replication factor of 1 to ensure that the scheduling service is always available.

   Here's an example Docker Stack file (`swarm-cron-stack.yml`) to deploy `swarm-cron`:

   ```yaml
   version: '3.7'

   services:
     swarm-cron:
       image: anboo/swarm-cron:latest
       volumes:
         - /var/run/docker.sock:/var/run/docker.sock
       deploy:
         placement:
           constraints: [node.role == manager]
         replicas: 1
       environment:
         - LOG_LEVEL=info  # Optional: Set the log level (debug, info, warn, error)
    ```

    Deploy the stack using the following command:
    
        docker stack deploy -c swarm-cron-stack.yml swarm_cron
2. **Schedule a Service:**

   To schedule a service to run at specific intervals, add the cronjob_schedule label to your service definition with a cron expression as its value. The cron expression defines the schedule on which the service will be executed.

   Example service definition with cronjob_schedule label:

   ```yaml
   version: '3.7'

   services:
     my-scheduled-task:
       image: my-task-image:latest
       deploy:
         labels:
           cronjob_schedule: "* * * * * *"  # Runs every minute
   ```