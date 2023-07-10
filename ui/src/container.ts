import { LitElement, html, css } from "lit";
import { customElement, state } from "lit/decorators.js";
import report from "./report.json";
import { Nav } from "./nav-header";
import "./nav-header";
import "./export-items";
import "./dead-files";

@customElement("app-container")
export class Container extends LitElement {
  static styles = css``;

  @state()
  active = Nav.all;

  onNav = (nav: Nav) => {
    this.active = nav;
  };

  render() {
    console.log(report);
    let content = html`<export-items
      .exports="${report.exports}"
    ></export-items>`;
    switch (this.active) {
      case Nav.dead: {
        content = html`<dead-files
          .dead_files="${report.dead_files}"
        ></dead-files>`;
        break;
      }
      case Nav.pJson: {
        content = html`<div>Package JSON</div>`;
        break;
      }
    }
    return html`
      <div>
        <nav-header .onClick="${this.onNav}"></nav-header>
        ${content}
      </div>
    `;
  }
}
