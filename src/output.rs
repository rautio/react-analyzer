use super::extractor::Output;
use serde_json;
use std::fs::File;
use std::io::BufWriter;
use std::io::Write;

pub fn write_output(output: Output) -> std::io::Result<()> {
    let file = File::create("report.json")?;
    let mut writer = BufWriter::new(file);
    serde_json::to_writer(&mut writer, &serde_json::to_string(&output).unwrap())?;
    writer.flush().unwrap();
    Ok(())
}
