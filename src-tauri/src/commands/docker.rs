use crate::modules::docker::DockerClient;
use crate::error::Result;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct DockerStatus {
    pub connected: bool,
    pub version: Option<String>,
    pub error: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ImageInfo {
    pub id: String,
    pub tags: Vec<String>,
    pub size: i64,
    pub created: i64,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ContainerInfo {
    pub id: String,
    pub names: Vec<String>,
    pub image: String,
    pub state: String,
    pub status: String,
}

#[tauri::command]
pub async fn check_docker() -> Result<DockerStatus> {
    match DockerClient::new().await {
        Ok(docker) => {
            match docker.ping().await {
                Ok(_) => Ok(DockerStatus {
                    connected: true,
                    version: Some("connected".to_string()),
                    error: None,
                }),
                Err(e) => Ok(DockerStatus {
                    connected: false,
                    version: None,
                    error: Some(e.to_string()),
                }),
            }
        }
        Err(e) => Ok(DockerStatus {
            connected: false,
            version: None,
            error: Some(e.to_string()),
        }),
    }
}

#[tauri::command]
pub async fn list_images() -> Result<Vec<ImageInfo>> {
    let docker = DockerClient::new().await?;
    let images = docker.list_images().await?;

    Ok(images.into_iter().map(|i| {
        ImageInfo {
            id: i.id,
            tags: i.repo_tags,
            size: i.size,
            created: i.created,
        }
    }).collect())
}

#[tauri::command]
pub async fn list_containers() -> Result<Vec<ContainerInfo>> {
    let docker = DockerClient::new().await?;
    let containers = docker.list_containers().await?;

    Ok(containers.into_iter().map(|c| {
        ContainerInfo {
            id: c.id,
            names: c.names,
            image: c.image,
            state: c.state,
            status: c.status,
        }
    }).collect())
}
