use super::extract::Output;
use handlebars::Handlebars;
use serde_json;
use serde_json::json;
use std::fs::create_dir_all;
use std::fs::File;
use std::io::BufWriter;
use std::io::Write;
use std::path::Path;

pub fn write_output(output: &Output) -> std::io::Result<()> {
    // Write main report
    let mut file = File::create("ui/src/report.json")?;
    file.flush()?;
    let mut writer = BufWriter::new(file);
    serde_json::to_writer(&mut writer, &output)?;
    writer.flush().unwrap();

    // Write individual file templates
    let mut handlebars = Handlebars::new();
    // Register template for files
    handlebars
        .register_template_file("template", "./src/templates/file.hbs")
        .unwrap();
    for node in &output.exports {
        // Construct file path
        let mut out_path = "./target/report/".to_string();
        out_path.push_str(node.source.as_str());
        out_path.push_str(r".html");
        // Create directories needed
        let path = Path::new(&out_path);
        create_dir_all(path.parent().unwrap())?;
        // Create file
        let mut file = File::create(out_path)?;
        file.flush()?;
        // Write template
        let _ = handlebars.render_to_write(
            "template",
            &json!({"file_name": node.source, "exports": node.exports}),
            &mut file,
        );
    }
    Ok(())
}
