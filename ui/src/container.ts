import { LitElement, html, css } from "lit";
import { customElement, state } from "lit/decorators.js";
import report from "./report.json";
import { Nav } from "./nav-header";
import "./nav-header";
import "./export-items";
import "./dead-files";
import "./unknown-imports";
import "./package-json";

@customElement("app-container")
export class Container extends LitElement {
  static styles = css`
    .wrapper {
      margin: 20px;
    }
  `;

  @state()
  active = Nav.all;

  onNav = (nav: Nav) => {
    this.active = nav;
  };

  render() {
    console.log(report);
    // @ts-ignore
    window.report = report;
    let content = html`<export-items
      .exports="${report.exports}"
      .summary="${report.summary}"
    ></export-items>`;
    switch (this.active) {
      case Nav.dead: {
        content = html`<dead-files
          .dead_files="${report.dead_files}"
        ></dead-files>`;
        break;
      }
      case Nav.unknown: {
        content = html`<unknown-imports
          .unknown_imports="${report.unknown_imports}"
        ></unknown-imports>`;
        break;
      }
      case Nav.pJson: {
        content = html`<package-json
          .package_json="${report.package_json}"
        ></package-json>`;
        break;
      }
    }
    return html`
      <div>
        <nav-header
          .active="${this.active}"
          .onClick="${this.onNav}"
        ></nav-header>
        <div class="wrapper">${content}</div>
      </div>
    `;
  }
}
