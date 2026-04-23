use crate::modules::config::{ConfigManager, ProjectConfig};
use crate::error::Result;
use chrono::Utc;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Serialize, Deserialize)]
pub struct Project {
    pub id: String,
    pub name: String,
    pub path: String,
    pub language: String,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct CreateProjectRequest {
    pub name: String,
    pub path: String,
    pub language: String,
}

#[tauri::command]
pub async fn list_projects() -> Result<Vec<Project>> {
    let config_manager = ConfigManager::new();
    let config = config_manager.load_config()?;

    Ok(config.projects.into_iter().map(|p| Project {
        id: p.id,
        name: p.name,
        path: p.path,
        language: p.language,
        created_at: p.created_at.to_rfc3339(),
        updated_at: p.updated_at.to_rfc3339(),
    }).collect())
}

#[tauri::command]
pub async fn create_project(request: CreateProjectRequest) -> Result<Project> {
    let config_manager = ConfigManager::new();

    let project = ProjectConfig {
        id: Uuid::new_v4().to_string(),
        name: request.name.clone(),
        path: request.path.clone(),
        language: request.language.clone(),
        build_config: None,
        deploy_config: None,
        created_at: Utc::now(),
        updated_at: Utc::now(),
    };

    config_manager.add_project(project.clone())?;

    Ok(Project {
        id: project.id,
        name: project.name,
        path: project.path,
        language: project.language,
        created_at: project.created_at.to_rfc3339(),
        updated_at: project.updated_at.to_rfc3339(),
    })
}

#[tauri::command]
pub async fn get_project(project_id: String) -> Result<Option<Project>> {
    let config_manager = ConfigManager::new();
    let config = config_manager.load_config()?;

    let project = config.projects.into_iter()
        .find(|p| p.id == project_id)
        .map(|p| Project {
            id: p.id,
            name: p.name,
            path: p.path,
            language: p.language,
            created_at: p.created_at.to_rfc3339(),
            updated_at: p.updated_at.to_rfc3339(),
        });

    Ok(project)
}

#[tauri::command]
pub async fn update_project(project_id: String, name: String, path: String, language: String) -> Result<Project> {
    let config_manager = ConfigManager::new();
    let mut config = config_manager.load_config()?;

    let project = config.projects.iter_mut()
        .find(|p| p.id == project_id)
        .ok_or_else(|| crate::error::FlowCIError::Project("Project not found".to_string()))?;

    project.name = name.clone();
    project.path = path.clone();
    project.language = language.clone();
    project.updated_at = Utc::now();

    let updated_project = project.clone();
    config_manager.save_config(&config)?;

    Ok(Project {
        id: updated_project.id,
        name: updated_project.name,
        path: updated_project.path,
        language: updated_project.language,
        created_at: updated_project.created_at.to_rfc3339(),
        updated_at: updated_project.updated_at.to_rfc3339(),
    })
}

#[tauri::command]
pub async fn delete_project(project_id: String) -> Result<()> {
    let config_manager = ConfigManager::new();
    config_manager.remove_project(&project_id)
}
