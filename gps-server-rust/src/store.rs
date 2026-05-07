use std::collections::HashSet;
use tokio::sync::Mutex;

lazy_static::lazy_static! {
    pub static ref DEVICES: Mutex<HashSet<String>> = {
        let mut set = HashSet::new();
        set.insert("359339075496001".to_string()); // your device IMEI
        Mutex::new(set)
    };
}