use crate::get_api_base_url;
use reqwest::Client;
use serde::de::DeserializeOwned;
use serde::Serialize;

pub struct HttpClient {
    client: Client,
    base_url: String,
}

impl HttpClient {
    pub fn new() -> Self {
        Self {
            client: Client::new(),
            base_url: get_api_base_url(),
        }
    }

    pub async fn get<T: DeserializeOwned>(&self, path: &str) -> Result<T, String> {
        let url = format!("{}{}", self.base_url, path);
        self.client
            .get(&url)
            .send()
            .await
            .map_err(|e| e.to_string())?
            .json::<T>()
            .await
            .map_err(|e| e.to_string())
    }

    pub async fn post<T: DeserializeOwned, B: Serialize>(&self, path: &str, body: &B) -> Result<T, String> {
        let url = format!("{}{}", self.base_url, path);
        self.client
            .post(&url)
            .json(body)
            .send()
            .await
            .map_err(|e| e.to_string())?
            .json::<T>()
            .await
            .map_err(|e| e.to_string())
    }

    pub async fn put<T: DeserializeOwned, B: Serialize>(&self, path: &str, body: &B) -> Result<T, String> {
        let url = format!("{}{}", self.base_url, path);
        self.client
            .put(&url)
            .json(body)
            .send()
            .await
            .map_err(|e| e.to_string())?
            .json::<T>()
            .await
            .map_err(|e| e.to_string())
    }

    pub async fn delete<T: DeserializeOwned>(&self, path: &str) -> Result<T, String> {
        let url = format!("{}{}", self.base_url, path);
        self.client
            .delete(&url)
            .send()
            .await
            .map_err(|e| e.to_string())?
            .json::<T>()
            .await
            .map_err(|e| e.to_string())
    }
}

impl Default for HttpClient {
    fn default() -> Self {
        Self::new()
    }
}
