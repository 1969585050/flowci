pub mod error;

pub mod commands {
    pub mod build;
    pub mod deploy;
    pub mod docker;
    pub mod http_client;
    pub mod project;
    pub mod push;
}

use std::env;

pub fn get_api_base_url() -> String {
    env::var("FLOWCI_API_URL").unwrap_or_else(|_| "http://localhost:3847".to_string())
}

pub fn init_logging() {
    tracing_subscriber::registry()
        .with(tracing_subscriber::fmt::layer())
        .with(tracing_subscriber::EnvFilter::from_default_env()
            .add_directive("flowci=info".parse().unwrap()))
        .init();
}

pub fn run() {
    init_logging();
    tracing::info!("FlowCI starting...");
    tracing::info!("API URL: {}", get_api_base_url());

    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_dialog::init())
        .invoke_handler(tauri::generate_handler![
            commands::project::list_projects,
            commands::project::create_project,
            commands::project::get_project,
            commands::project::update_project,
            commands::project::delete_project,
            commands::build::build_image,
            commands::build::get_build_logs,
            commands::build::list_builds,
            commands::push::push_image,
            commands::deploy::deploy_compose,
            commands::deploy::get_deploy_status,
            commands::deploy::rollback_deploy,
            commands::docker::check_docker,
            commands::docker::list_images,
            commands::docker::list_containers,
        ])
        .setup(|_app| {
            tracing::info!("FlowCI setup complete");
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
