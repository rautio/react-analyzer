import { LitElement, html, css } from "lit";
import { customElement, state } from "lit/decorators.js";
import { Nav } from "./nav-header";
import './nav-header';
import './export-items';


@customElement("app-container")
export class Container extends LitElement {
  static styles = css``;

  @state()
  active = Nav.all;

  onNav = (nav: Nav) => {
    this.active = nav;
  };

  render() {
    return html`
    <div>
      <nav-header .onClick="${this.onNav}"></nav-header>
      <export-items></export-items>
    </div>
    `;
  }
}
