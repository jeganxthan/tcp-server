pub fn parse_imei(data: &[u8]) -> Option<String> {
    if data.len() < 12 {
        return None;
    }

    let imei_bytes = &data[4..12];
    Some(hex::encode(imei_bytes))
}

pub fn get_protocol(data: &[u8]) -> Option<u8> {
    if data.len() > 3 {
        Some(data[3])
    } else {
        None
    }
}