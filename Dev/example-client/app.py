"""Example client for pvauto.linhazero.com API — Streamlit UI em português brasileiro."""

import json
import os
import re
from urllib.parse import urlencode, urlparse, parse_qs

import requests
import streamlit as st

BASE_URL = os.environ.get("PV_API_BASE_URL", "https://pvauto.linhazero.com")
_DEFAULT_API_TOKEN = os.environ.get("PV_API_TOKEN", "")

# ── Session state defaults ──────────────────────────────────────────────────

_DEFAULTS = {
    "auth_token": None,
    "token_products": None,
    "token_retiradas": None,
    "token_selected": {},
    "token_personalizacoes": {},
}
for key, val in _DEFAULTS.items():
    if key not in st.session_state:
        st.session_state[key] = val


# ── Helpers ─────────────────────────────────────────────────────────────────

def req(method: str, path: str, **kwargs):
    url = f"{BASE_URL.rstrip('/')}/{path.lstrip('/')}"
    headers = kwargs.pop("headers", {})
    if method.upper() in ("POST", "PUT", "PATCH"):
        headers.setdefault("Content-Type", "application/json")
    token = kwargs.pop("token", None)
    if token:
        headers["Authorization"] = f"Bearer {token}"
    data = kwargs.pop("data", None)
    if data is not None and isinstance(data, dict):
        kwargs["data"] = json.dumps(data)
    return requests.request(method, url, headers=headers, **kwargs)


def fmt_json(r: requests.Response):
    try:
        return json.dumps(r.json(), indent=2, ensure_ascii=False)
    except Exception:
        return r.text


def fmt_status(r: requests.Response):
    emoji = "✅" if r.ok else "❌"
    return f"{emoji} **Status:** {r.status_code}"


def show_response(r: requests.Response):
    st.markdown(fmt_status(r))
    st.code(fmt_json(r), language="json")


# ── Page config ─────────────────────────────────────────────────────────────

st.set_page_config(
    page_title="pvauto — Cliente de Exemplo",
    page_icon="🛒",
    layout="centered",
)

st.title("🛒 pvauto — Cliente de Exemplo")
st.markdown("Interface de demonstração para a API do **pvauto.linhazero.com**")

# ── Sidebar navigation ──────────────────────────────────────────────────────

menu = st.sidebar.radio("Menu", ["Checkout via Token", "Requisição HTTP"])

# ─────────────────────────────────────────────────────────────────────────────
# CHECKOUT VIA TOKEN
# ─────────────────────────────────────────────────────────────────────────────

if menu == "Checkout via Token":
    st.header("🪙 Checkout via Token de Remessa")
    st.caption("Fluxo: token de remessa → produtos (id = package_id) + retiradas → confirmar compra.")

    # ── Section 0: API Key ────────────────────────────────────────────────
    st.subheader("0. Chave de API (API Token)")
    st.caption("O endpoint /checkouts/confirm requer autenticação via API token (Bearer pv_...).")

    api_token = st.text_input(
        "API Token (Bearer)",
        value=st.session_state.get("token_api_token", ""),
        placeholder="pv_... (deixe vazio para usar chave integrada)",
        key="token_api_token",
    )

    # ── Section 1: Token de Remessa ───────────────────────────────────────
    st.subheader("1. Token de Remessa")

    col_tok, col_load = st.columns([3, 1])
    with col_tok:
        remessa_token = st.text_input(
            "Token de remessa",
            value="",
            placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
            key="remessa_token",
        )
    with col_load:
        st.write("")
        if st.button("📦 Carregar", key="load_token_btn", use_container_width=True):
            if not remessa_token.strip():
                st.error("Informe o token de remessa.")
            else:
                with st.spinner("Carregando produtos..."):
                    r = req("GET", "/produtos", params={"token": remessa_token.strip()})
                if r.ok:
                    data = r.json()
                    st.session_state.token_products = data.get("produtos", [])
                    st.session_state.token_retiradas = data.get("retiradas", [])
                    st.session_state.token_selected = {}
                    st.session_state.token_personalizacoes = {}
                    st.success(f"{len(st.session_state.token_products)} produtos carregados.")
                else:
                    st.session_state.token_products = []
                    st.session_state.token_retiradas = []
                    show_response(r)

    # ── Section 2: Retiradas ──────────────────────────────────────────────
    retiradas = st.session_state.get("token_retiradas", [])
    if retiradas:
        st.subheader("2. Pontos de retirada")
        for ret in retiradas:
            name = ret.get("name", "?")
            desc = ret.get("description", "")
            valid = ret.get("validUntil", "")
            addr = ret.get("address") or {}
            parts = [
                addr.get("street", ""),
                addr.get("number", ""),
                addr.get("neighborhood", ""),
                addr.get("city", ""),
                addr.get("state", ""),
            ]
            addr_str = ", ".join(p for p in parts if p)
            extra = []
            if desc:
                extra.append(desc)
            if valid:
                extra.append(f"Válido até: {valid[:10]}")
            text = f"📍 {name}  (ID {ret.get('id')})  —  {addr_str}"
            if extra:
                text += "\n" + " — ".join(extra)
            st.code(text, language=None)

    # ── Section 3: Produtos ───────────────────────────────────────────────
    produtos = st.session_state.get("token_products", [])
    if produtos:
        st.markdown("---")
        st.subheader("3. Selecionar produtos")
        st.caption("O id de cada produto é o package_id para o checkout. Escolha quantidades e, opcionalmente, a retirada e personalização.")

        # Build retirada lookup
        retirada_by_id = {r["id"]: r for r in retiradas} if retiradas else {}

        for prod in produtos:
            pkg_id = prod.get("id")
            name = prod.get("name", "?")
            price = prod.get("price", 0)
            discount = prod.get("discount", 0)
            discount_type = prod.get("discount_type", "")
            stock = prod.get("stock", 0)

            cur_qty = st.session_state.get("token_selected", {}).get(pkg_id, 0)

            with st.expander(f"{name} (ID {pkg_id}) — R$ {price} — estoque: {stock}", expanded=cur_qty > 0):
                # Product info
                col_a, col_b = st.columns(2)
                with col_a:
                    st.markdown(f"**Preço:** R$ {price}")
                    if discount:
                        disc_label = f"{discount}%" if discount_type == "percentage" else f"R$ {discount}"
                        st.markdown(f"**Desconto:** {disc_label} ({discount_type})")
                    st.markdown(f"**Código:** {prod.get('code', '-')}")
                    st.markdown(f"**Categoria:** {prod.get('category', '-')} (ID {prod.get('category_id', '-')})")
                    st.markdown(f"**Destaque:** {'Sim' if prod.get('featured') else 'Não'}")
                with col_b:
                    st.markdown(f"**UUID:** `{prod.get('uuid', '-')}`")
                    colors = prod.get("colors") or []
                    st.markdown(f"**Cores:** {', '.join(colors) if colors else '-'}")
                    sizes = prod.get("sizes") or []
                    st.markdown(f"**Tamanhos:** {', '.join(sizes) if sizes else '-'}")
                    st.markdown(f"**Personalização:** nome={'Sim' if prod.get('allowCustomName') else 'Não'}, modalidade={'Sim' if prod.get('allowCustomModality') else 'Não'}, cor={'Sim' if prod.get('allowCustomColorSelection') else 'Não'}")

                # Description
                desc = prod.get("description", "")
                if desc:
                    st.markdown(f"**Descrição:** {desc}")

                # Specifications
                specs = prod.get("specifications") or []
                if specs:
                    st.markdown("**Especificações:**")
                    st.dataframe(
                        [{"Nome": s.get("name", ""), "Valor": s.get("value", "")} for s in specs],
                        use_container_width=True,
                        hide_index=True,
                    )

                # Image
                img = prod.get("imageUrl", "")
                if img:
                    st.image(img, width=200)

                # Extra images (collapsed by default)
                spec_imgs = prod.get("specificationsImages") or []
                desc_imgs = prod.get("descriptionImages") or []
                deli_imgs = prod.get("deliveryImages") or []
                if spec_imgs or desc_imgs or deli_imgs:
                    st.caption("📷 Mais imagens")
                    for url in spec_imgs:
                        st.caption("Especificação")
                        st.image(url, width=200)
                    for url in desc_imgs:
                        st.caption("Descrição")
                        st.image(url, width=200)
                    for url in deli_imgs:
                        st.caption("Entrega")
                        st.image(url, width=200)

                # Quantity
                qty_key = f"tkn_qty_{pkg_id}"
                new_qty = st.number_input(
                    "Quantidade",
                    min_value=0,
                    max_value=stock,
                    value=cur_qty,
                    key=qty_key,
                )
                if new_qty != cur_qty:
                    if "token_selected" not in st.session_state:
                        st.session_state.token_selected = {}
                    if new_qty > 0:
                        st.session_state.token_selected[pkg_id] = new_qty
                    elif pkg_id in st.session_state.token_selected:
                        del st.session_state.token_selected[pkg_id]

                # Retirada selection (only if there are retiradas and qty > 0)
                if retiradas and new_qty > 0:
                    ret_options = {f"{r['name']} (ID {r['id']})": r["id"] for r in retiradas}
                    ret_key = f"tkn_ret_{pkg_id}"
                    cur_ret = st.session_state.get("token_retiradas_selected", {}).get(pkg_id, retiradas[0]["id"])
                    selected_ret = st.selectbox(
                        "Ponto de retirada",
                        list(ret_options.keys()),
                        index=list(ret_options.values()).index(cur_ret) if cur_ret in ret_options.values() else 0,
                        key=ret_key,
                    )
                    if "token_retiradas_selected" not in st.session_state:
                        st.session_state.token_retiradas_selected = {}
                    st.session_state.token_retiradas_selected[pkg_id] = ret_options[selected_ret]

                # Personalizacao
                if new_qty > 0:
                    st.markdown("**Personalização (opcional):**")
                    pers_key = f"tkn_pers_{pkg_id}"
                    if pers_key not in st.session_state.token_personalizacoes:
                        st.session_state.token_personalizacoes[pers_key] = []

                    existing = st.session_state.token_personalizacoes[pers_key]

                    # Show existing entries
                    to_delete = None
                    for i, entry in enumerate(existing):
                        col_n, col_v, col_x = st.columns([2, 2, 1])
                        with col_n:
                            st.text_input("Nome", value=entry["nome"], key=f"{pers_key}_n{i}", disabled=True, label_visibility="collapsed")
                        with col_v:
                            st.text_input("Valor", value=entry["valor"], key=f"{pers_key}_v{i}", disabled=True, label_visibility="collapsed")
                        with col_x:
                            if st.button("✕", key=f"{pers_key}_x{i}"):
                                to_delete = i

                    if to_delete is not None:
                        del st.session_state.token_personalizacoes[pers_key][to_delete]
                        st.rerun()

                    # Add new entry
                    col_n, col_v, col_add = st.columns([2, 2, 1])
                    with col_n:
                        new_nome = st.text_input("Nome", key=f"{pers_key}_new_n", placeholder="Name", label_visibility="collapsed")
                    with col_v:
                        new_valor = st.text_input("Valor", key=f"{pers_key}_new_v", placeholder="Fulano", label_visibility="collapsed")
                    with col_add:
                        if st.button("＋", key=f"{pers_key}_add"):
                            if new_nome.strip() and new_valor.strip():
                                st.session_state.token_personalizacoes[pers_key].append({
                                    "nome": new_nome.strip(),
                                    "valor": new_valor.strip(),
                                })
                                st.rerun()

        if st.button("🧹 Limpar seleção", key="clear_token_selection"):
            st.session_state.token_selected = {}
            st.session_state.token_personalizacoes = {}
            st.session_state.token_retiradas_selected = {}
            st.rerun()
    else:
        st.info("Informe um token de remessa e clique em **Carregar**.")

    # ── Section 4: Confirmar ──────────────────────────────────────────────
    st.markdown("---")
    st.subheader("4. Confirmar compra")

    col_nome, _ = st.columns(2)
    with col_nome:
        client_name = st.text_input("Nome do cliente (opcional)", value="", placeholder="Fulano de Tal", key="tok_client_name")

    col_tel, col_cpf = st.columns(2)
    with col_tel:
        telephone = st.text_input(
            "Telefone (opcional)",
            value="",
            placeholder="11988887777",
            key="tok_telephone",
            help="Dígitos são extraídos; formatação é ignorada. Ao menos um de CPF ou telefone é obrigatório.",
        )
    with col_cpf:
        cpf = st.text_input(
            "CPF (opcional)",
            value="",
            placeholder="12345678901",
            key="tok_cpf",
            help="11 dígitos; formatação é ignorada. Tem prioridade sobre telefone.",
        )

    selected = st.session_state.get("token_selected", {})
    if selected:
        st.caption(
            f"{sum(selected.values())} itens em "
            f"{len(selected)} produtos selecionados."
        )

    if st.button("✅ Confirmar compra", type="primary", use_container_width=True, key="confirm_token_btn"):
        if not selected:
            st.error("Selecione pelo menos um produto (seção 3).")
        else:
            ret_selected = st.session_state.get("token_retiradas_selected", {})
            items = []
            for pkg_id, qty in selected.items():
                item = {"id": int(pkg_id), "quantity": qty}
                ret_id = ret_selected.get(pkg_id)
                if ret_id is not None:
                    item["retirada"] = ret_id
                pers = st.session_state.token_personalizacoes.get(f"tkn_pers_{pkg_id}", [])
                if pers:
                    item["personalizacao"] = pers
                items.append(item)

            payload = {"items": items}
            if client_name.strip():
                payload["client_name"] = client_name.strip()
            if telephone.strip():
                payload["telephone"] = telephone.strip()
            if cpf.strip():
                payload["cpf"] = cpf.strip()

            api_token_val = st.session_state.get("token_api_token", "").strip() or _DEFAULT_API_TOKEN
            if not api_token_val:
                st.error("Informe o API Token (seção 0).")
            else:
                with st.spinner("Finalizando compra..."):
                    r = req(
                        "POST",
                        "/checkouts/confirm",
                        data=payload,
                        token=api_token_val,
                    )
                if r.ok:
                    data = r.json()
                    unique_id = data.get("unique_id", "?")
                    status = data.get("status", "?")
                    qr_codes = data.get("qr_codes", [])
                    st.balloons()  # ponytail: deprecated since Streamlit 1.42, replace with st.toast or custom confetti if noise matters
                    st.success(
                        f"**Compra confirmada!**  \n"
                        f"ID único: `{unique_id}`  \n"
                        f"Status: `{status}`"
                    )
                    if qr_codes:
                        st.subheader("📱 QR Codes para retirada")
                        for entry in qr_codes:
                            st.markdown(f"**Produto ID {entry.get('product_id', '?')}** — Qtd: {entry.get('quantity', '?')} — R$ {entry.get('price', '?')}")
                            disc = entry.get("discount", 0)
                            disc_type = entry.get("discount_type", "")
                            if disc:
                                disc_label = f"{disc}%" if disc_type == "percentage" else f"R$ {disc}"
                                st.caption(f"Desconto: {disc_label}")
                            pers_out = entry.get("personalizacao") or []
                            if pers_out:
                                st.caption("Personalização: " + ", ".join(f"{p['nome']}={p['valor']}" for p in pers_out))
                            ret_out = entry.get("retirada") or {}
                            if ret_out:
                                addr_out = ret_out.get("address") or {}
                                addr_parts = [
                                    addr_out.get("street", ""),
                                    addr_out.get("number", ""),
                                    addr_out.get("neighborhood", ""),
                                    addr_out.get("city", ""),
                                    addr_out.get("state", ""),
                                ]
                                st.caption(f"Retirada: {ret_out.get('nome', '?')} — {', '.join(p for p in addr_parts if p)}")
                            qr_data_uri = entry.get("qr_code", "")
                            if qr_data_uri:
                                st.image(qr_data_uri, width=200)
                            st.markdown("---")
                    st.session_state.token_selected = {}
                    st.session_state.token_personalizacoes = {}
                    st.session_state.token_retiradas_selected = {}
                else:
                    show_response(r)

    # ── Section 5: Cancel checkout ────────────────────────────────────────
    st.markdown("---")
    st.subheader("5. Cancelar checkout")
    st.caption("Cancela um checkout previamente confirmado. Apenas pacotes PENDING são cancelados; RECEIVED permanecem.")

    cancel_unique_id = st.text_input(
        "UUID do checkout (unique_id)",
        value="",
        placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
        key="cancel_unique_id",
    )

    if st.button("🗑️ Cancelar checkout", use_container_width=True, key="cancel_checkout_btn"):
        api_token_val = st.session_state.get("token_api_token", "").strip() or _DEFAULT_API_TOKEN
        if not cancel_unique_id.strip():
            st.error("Informe o UUID do checkout.")
        else:
            with st.spinner("Cancelando checkout..."):
                r = req(
                    "DELETE",
                    f"/checkouts/{cancel_unique_id.strip()}",
                    token=api_token_val,
                )
            if r.ok:
                data = r.json()
                st.success(
                    f"**Checkout cancelado!**  \n"
                    f"Pacotes cancelados: `{data.get('cancelled_count', '?')}`  \n"
                    f"Pacotes restantes: `{data.get('remaining_count', '?')}`"
                )
                if data.get("remaining_count", 0) == 0:
                    st.info("Todos os pacotes foram cancelados — o checkout foi removido.")
            else:
                show_response(r)

# ─────────────────────────────────────────────────────────────────────────────
# REQUISIÇÃO HTTP
# ─────────────────────────────────────────────────────────────────────────────

elif menu == "Requisição HTTP":
    st.header("🌐 Requisição HTTP genérica")
    st.markdown("Envie requisições arbitrárias para qualquer endpoint da API.")

    col1, col2 = st.columns([1, 3])
    with col1:
        http_method = st.selectbox("Método", ["GET", "POST", "PUT", "DELETE", "PATCH"])
    with col2:
        endpoint = st.text_input(
            "Endpoint",
            value="/produtos",
            placeholder="/produtos",
            help="Caminho da API, ex: /produtos, /login, /checkouts/confirm",
        )

    use_auth = st.checkbox("Autenticar", value=False, key="req_auth")
    token_override = None
    if use_auth:
        col1, col2 = st.columns(2)
        with col1:
            auth_email = st.text_input("E-mail", key="req_email")
        with col2:
            auth_pass = st.text_input("Senha", type="password", key="req_pass")

    body_visible = http_method in ("POST", "PUT", "PATCH")
    body_text = ""
    if body_visible:
        body_text = st.text_area(
            "Corpo (JSON)",
            value='{"key": "value"}',
            height=150,
            key="req_body",
        )

    query_str = st.text_input(
        "Query string (opcional)",
        value="",
        placeholder="session_id=cart_123&foo=bar",
        key="req_query",
    )

    if st.button("🚀 Enviar", type="primary", use_container_width=True):
        if not endpoint.strip():
            st.error("Informe o endpoint.")
            st.stop()

        if use_auth:
            with st.spinner("Autenticando..."):
                r_auth = req("POST", "/login", data={"email": auth_email, "password": auth_pass})
            if r_auth.ok:
                data = r_auth.json()
                token_override = data.get("access_token") or data.get("token", "")
                if not token_override:
                    st.error("Falha ao obter token de autenticação.")
                    show_response(r_auth)
                    token_override = None
            else:
                st.error("Falha na autenticação.")
                show_response(r_auth)

        path = endpoint.strip().lstrip("/")
        params = None
        if query_str.strip():
            parsed = urlparse("?" + query_str.strip())
            params = {k: v[0] if len(v) == 1 else v for k, v in parse_qs(parsed.query).items()}

        data_payload = None
        if body_visible and body_text.strip():
            try:
                data_payload = json.loads(body_text.strip())
            except json.JSONDecodeError as e:
                st.error(f"JSON inválido no corpo: {e}")
                st.stop()

        with st.spinner(f"Enviando {http_method} {path}..."):
            r = req(
                http_method,
                path,
                params=params,
                data=data_payload,
                token=token_override,
            )
        show_response(r)
