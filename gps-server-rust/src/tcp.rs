use tokio::net::TcpListener;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use crate::parser::{parse_imei, get_protocol};
use crate::store::DEVICES;

pub async fn start_server() {
    let listener = TcpListener::bind("0.0.0.0:5023").await.unwrap();

    println!("🚀 TCP Server running on port 5023");

    loop {
        let (mut socket, addr) = listener.accept().await.unwrap();

        println!("✅ Connected: {}", addr);

        tokio::spawn(async move {
            let mut buffer = [0u8; 1024];

            loop {
                let n = match socket.read(&mut buffer).await {
                    Ok(n) if n == 0 => {
                        println!("❌ Disconnected: {}", addr);
                        return;
                    }
                    Ok(n) => n,
                    Err(_) => return,
                };

                let data = &buffer[..n];

                println!("📦 RAW: {}", hex::encode(data));

                if let Some(protocol) = get_protocol(data) {
                    match protocol {

                        // 🔐 Login Packet
                        0x01 => {
                            println!("🔐 Login packet");

                            if let Some(imei) = parse_imei(data) {
                                println!("📱 IMEI: {}", imei);

                                let devices = DEVICES.lock().unwrap();

                                if devices.contains(&imei) {
                                    println!("✅ Device authenticated");

                                    send_ack(&mut socket).await;
                                } else {
                                    println!("❌ Unknown device");
                                }
                            }
                        }

                        // 📍 GPS Packet
                        0x12 => {
                            println!("📍 GPS data received");
                        }

                        _ => {
                            println!("❓ Unknown protocol: {}", protocol);
                        }
                    }
                }
            }
        });
    }
}

async fn send_ack(socket: &mut tokio::net::TcpStream) {
    let ack = vec![
        0x78, 0x78,
        0x05,
        0x01,
        0x00, 0x01,
        0x0D, 0x0A,
    ];

    let _ = socket.write_all(&ack).await;
    println!("📤 ACK sent");
}