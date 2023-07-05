use super::extract::Output;
use serde_json;
use std::fs::File;
use std::io::BufWriter;
use std::io::Write;

pub fn write_output(output: Output) -> std::io::Result<()> {
    let file = File::create("ui/src/report.json")?;
    let mut writer = BufWriter::new(file);
    serde_json::to_writer(&mut writer, &output)?;
    writer.flush().unwrap();
    Ok(())
}
