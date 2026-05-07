mod tcp;
mod parser;
mod store;

#[tokio::main]
async fn main() {
    tcp::start_server().await;
}