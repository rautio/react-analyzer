import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";

export enum Nav {
  all,
  dead,
  pJson,
}

// type OnClick = (nav: Nav) => void;


@customElement("nav-header")
export class NavHeader extends LitElement {
  static styles = css``;

  @property({ type: Nav })
  active = Nav.all

  // @property({ type: OnClick })
  @property()
  onClick = () => {}
  constructor() {
    super();
  }

  _handleClick = (nav: Nav) => {
    // @ts-ignore
    this.onClick(nav);
  };

  render() {
    return html`
      <nav>
        <button @click="${() => this._handleClick(Nav.all)}" class="nav-item">All Files</button>
        <button @click="${() => this._handleClick(Nav.dead)}" class="nav-item">Dead Files</button>
        <button @click="${() => this._handleClick(Nav.pJson)}" class="nav-item">Package JSON</button>
      </nav>
    `;
  }
}
