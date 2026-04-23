use crate::modules::deployer::{Deployer, DeployConfig, DeploymentType};
use crate::error::Result;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct DeployRequest {
    pub project_id: String,
    pub image_tag: String,
    pub deployment_type: String,
    pub environment_vars: std::collections::HashMap<String, String>,
    pub replicas: Option<u32>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DeployResponse {
    pub id: String,
    pub status: String,
    pub containers: Vec<ContainerResponse>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ContainerResponse {
    pub name: String,
    pub image: String,
    pub status: String,
    pub ports: Vec<PortResponse>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct PortResponse {
    pub host_port: u16,
    pub container_port: u16,
    pub protocol: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DeployStatusResponse {
    pub project_id: String,
    pub status: String,
    pub containers: Vec<ContainerResponse>,
}

#[tauri::command]
pub async fn deploy_compose(request: DeployRequest) -> Result<DeployResponse> {
    let deployer = Deployer::new();

    let config = DeployConfig {
        project_id: request.project_id.clone(),
        deployment_type: match request.deployment_type.as_str() {
            "compose" | "docker-compose" => DeploymentType::DockerCompose,
            "rolling" | "rolling_update" => DeploymentType::RollingUpdate,
            _ => DeploymentType::SingleContainer,
        },
        compose_file: None,
        environment_vars: request.environment_vars,
        image_tag: request.image_tag.clone(),
        replicas: request.replicas,
    };

    let compose_content = deployer.generate_compose_file(&config, &request.project_id)?;
    let result = deployer.deploy_compose(&compose_content, &request.project_id).await?;

    Ok(DeployResponse {
        id: result.id,
        status: format!("{:?}", result.status).to_lowercase(),
        containers: result.containers.into_iter().map(|c| {
            ContainerResponse {
                name: c.name,
                image: c.image,
                status: c.status,
                ports: c.ports.into_iter().map(|p| {
                    PortResponse {
                        host_port: p.host_port,
                        container_port: p.container_port,
                        protocol: p.protocol,
                    }
                }).collect(),
            }
        }).collect(),
    })
}

#[tauri::command]
pub async fn get_deploy_status(project_id: String) -> Result<DeployStatusResponse> {
    let deployer = Deployer::new();
    let containers = deployer.get_deploy_status(&project_id).await?;

    Ok(DeployStatusResponse {
        project_id,
        status: if containers.is_empty() { "not_deployed".to_string() } else { "running".to_string() },
        containers: containers.into_iter().map(|c| {
            ContainerResponse {
                name: c.name,
                image: c.image,
                status: c.status,
                ports: c.ports.into_iter().map(|p| {
                    PortResponse {
                        host_port: p.host_port,
                        container_port: p.container_port,
                        protocol: p.protocol,
                    }
                }).collect(),
            }
        }).collect(),
    })
}

#[tauri::command]
pub async fn rollback_deploy(project_id: String) -> Result<DeployResponse> {
    let deployer = Deployer::new();
    let result = deployer.rollback(&project_id).await?;

    Ok(DeployResponse {
        id: result.id,
        status: format!("{:?}", result.status).to_lowercase(),
        containers: vec![],
    })
}
