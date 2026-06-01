"""Example client for pvauto.linhazero.com API — Streamlit UI em português brasileiro."""

import json
import os
import re
import time
from urllib.parse import urlencode, urlparse, parse_qs

import requests
import streamlit as st

BASE_URL = os.environ.get("PV_API_BASE_URL", "https://pvauto.linhazero.com")

# ── Session state defaults ──────────────────────────────────────────────────

_DEFAULTS = {
    "cart_session_id": f"cart_{int(time.time())}",
    "auth_token": None,
    "products": None,
    "cart_items": None,
    "packages": None,
    "selected_package_id": None,
    "bank_account_id": None,
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


def sanitize_session_id(sid: str) -> str:
    return re.sub(r"[^a-zA-Z0-9_]", "", sid)


# ── Page config ─────────────────────────────────────────────────────────────

st.set_page_config(
    page_title="pvauto — Cliente de Exemplo",
    page_icon="🛒",
    layout="centered",
)

st.title("🛒 pvauto — Cliente de Exemplo")
st.markdown("Interface de demonstração para a API do **pvauto.linhazero.com**")

# ── Sidebar navigation ──────────────────────────────────────────────────────

menu = st.sidebar.radio("Menu", ["Checkout", "Banco & Saque", "Monitores", "Requisição HTTP"])

# ─────────────────────────────────────────────────────────────────────────────
# CHECKOUT
# ─────────────────────────────────────────────────────────────────────────────

if menu == "Checkout":
    st.header("🛍️ Checkout")

    # ── Session ID ──────────────────────────────────────────────────────────
    st.subheader("Sessão do carrinho")
    sid = st.text_input(
        "ID da sessão",
        value=st.session_state.cart_session_id,
        help="Identificador único do carrinho. Gerado automaticamente.",
    )
    st.session_state.cart_session_id = sanitize_session_id(sid)

    # ── Section 1: Products ────────────────────────────────────────────────
    st.subheader("1. Produtos disponíveis")
    if st.button("📦 Carregar produtos"):
        with st.spinner("Carregando produtos..."):
            r = req("GET", "/produtos")
            if r.ok:
                data = r.json()
                st.session_state.products = data if isinstance(data, list) else data.get("produtos", [])
                st.success(f"{len(st.session_state.products)} produtos carregados.")
            else:
                st.session_state.products = []
                show_response(r)

    if st.session_state.products:
        st.dataframe(
            [
                {
                    "ID": p.get("id"),
                    "Nome": p.get("nome") or p.get("name", ""),
                    "Preço": p.get("preco") or p.get("price", ""),
                }
                for p in st.session_state.products
            ],
            use_container_width=True,
        )
    else:
        st.info("Clique em **Carregar produtos** para ver a lista.")

    # ── Section 2: Cart ────────────────────────────────────────────────────
    st.subheader("2. Carrinho")
    col1, col2, col3 = st.columns([2, 1, 1])
    with col1:
        add_product_id = st.number_input("ID do produto", min_value=1, value=1, key="add_pid")
    with col2:
        add_qty = st.number_input("Quantidade", min_value=1, value=1, key="add_qty")
    with col3:
        st.write("")  # spacer
        st.write("")
        if st.button("➕ Adicionar ao carrinho", use_container_width=True):
            payload = {
                "product_id": int(add_product_id),
                "quantity": int(add_qty),
                "session_id": st.session_state.cart_session_id,
                "personalization": {},
            }
            with st.spinner("Adicionando..."):
                r = req("POST", "/carrinho/adicionar", data=payload)
            if r.ok:
                st.success("Produto adicionado ao carrinho!")
                st.session_state.cart_items = None  # force reload
            else:
                show_response(r)

    if st.button("🛒 Ver carrinho"):
        with st.spinner("Carregando carrinho..."):
            r = req("GET", f"/carrinho/dados?session_id={st.session_state.cart_session_id}")
        if r.ok:
            data = r.json()
            items = data if isinstance(data, list) else data.get("items", [])
            st.session_state.cart_items = items
        else:
            st.session_state.cart_items = []
            show_response(r)

    if st.session_state.cart_items:
        st.dataframe(
            [
                {
                    "Item ID": it.get("id") or it.get("item_id", ""),
                    "Produto": it.get("nome") or it.get("name") or it.get("product_name", ""),
                    "Quantidade": it.get("quantity", it.get("quantidade", "")),
                }
                for it in st.session_state.cart_items
            ],
            use_container_width=True,
        )
    else:
        st.info("Adicione produtos ao carrinho e clique em **Ver carrinho**.")

    # ── Section 3: Packages ────────────────────────────────────────────────
    st.subheader("3. Pacotes de retirada")
    if st.button("📋 Carregar pacotes"):
        with st.spinner("Carregando pacotes..."):
            r = req("GET", f"/carrinho/pacotes?session_id={st.session_state.cart_session_id}")
        if r.ok:
            data = r.json()
            products_data = data if isinstance(data, list) else data.get("products", [])
            all_pkgs = []
            for prod in products_data:
                for pkg in prod.get("packages", []):
                    pickup = pkg.get("pickup", {})
                    if isinstance(pickup, dict):
                        pkg["_pickup_name"] = pickup.get("name", pickup.get("nome", ""))
                    else:
                        pkg["_pickup_name"] = str(pickup)
                    all_pkgs.append(pkg)
            st.session_state.packages = all_pkgs
        else:
            st.session_state.packages = []
            show_response(r)

    if st.session_state.packages:
        pkg_options = {}
        for pkg in st.session_state.packages:
            pid = pkg.get("package_id") or pkg.get("id")
            price = pkg.get("price") or pkg.get("preco", "?")
            stock = pkg.get("stock") or pkg.get("estoque", "?")
            pickup_name = pkg.get("_pickup_name", "?")
            label = f"ID {pid} — R$ {price} — estoque: {stock} — retirada: {pickup_name}"
            pkg_options[label] = pid

        selected_label = st.radio(
            "Selecione um pacote",
            list(pkg_options.keys()),
            index=None,
        )
        if selected_label:
            st.session_state.selected_package_id = pkg_options[selected_label]
    else:
        st.info("Clique em **Carregar pacotes** para ver as opções.")

    # ── Section 4: Confirm ─────────────────────────────────────────────────
    st.subheader("4. Confirmar compra")
    col1, col2 = st.columns(2)
    with col1:
        value_paid = st.text_input("Valor pago (R$)", value="", placeholder="ex: 6.00")
    with col2:
        ext_ref = st.text_input("Referência externa", value="", placeholder="ex: EXT-123")

    col_q1, col_q2 = st.columns(2)
    with col_q1:
        confirm_qty = st.number_input("Quantidade", min_value=1, value=1, key="confirm_qty")
    with col_q2:
        st.write("")

    col3, col4 = st.columns(2)
    with col3:
        client_email = st.text_input(
            "E-mail do cliente (opcional)",
            value="",
            placeholder="cliente@exemplo.com",
            help="Se informado, cria conta de cliente e envia e-mail de boas-vindas.",
        )
    with col4:
        client_name = st.text_input("Nome do cliente (opcional)", value="", placeholder="Fulano de Tal")

    if st.button("✅ Confirmar compra", type="primary", use_container_width=True):
        if not value_paid:
            st.error("Informe o valor pago.")
        elif st.session_state.selected_package_id is None:
            st.error("Selecione um pacote de retirada (seção 3).")
        else:
            items = [
                {
                    "package_id": st.session_state.selected_package_id,
                    "quantity": confirm_qty,
                }
            ]
            payload = {
                "value_paid": value_paid,
                "external_reference": ext_ref or f"EXT-{int(time.time())}",
                "items": items,
            }
            if client_email.strip():
                payload["client_email"] = client_email.strip()
                if client_name.strip():
                    payload["client_name"] = client_name.strip()

            with st.spinner("Finalizando compra..."):
                r = req(
                    "POST",
                    f"/checkouts/confirm?session_id={st.session_state.cart_session_id}",
                    data=payload,
                )
            if r.ok:
                data = r.json()
                unique_id = data.get("unique_id", "?")
                status = data.get("status", "?")
                st.balloons()
                st.success(f"**Compra confirmada!**  \nID único: `{unique_id}`  \nStatus: `{status}`")
                # new session for next purchase
                st.session_state.cart_session_id = f"cart_{int(time.time())}"
                st.session_state.cart_items = None
                st.session_state.packages = None
                st.session_state.selected_package_id = None
            else:
                show_response(r)

# ─────────────────────────────────────────────────────────────────────────────
# BANCO & SAQUE
# ─────────────────────────────────────────────────────────────────────────────

elif menu == "Banco & Saque":
    st.header("🏦 Banco & Saque")

    # ── Section 1: Login ───────────────────────────────────────────────────
    st.subheader("1. Login (estabelecimento)")
    col1, col2, col3 = st.columns([2, 2, 1])
    with col1:
        login_email = st.text_input("E-mail", value="", key="bs_email")
    with col2:
        login_pass = st.text_input("Senha", type="password", value="", key="bs_pass")
    with col3:
        st.write("")
        st.write("")
        if st.button("🔑 Entrar", use_container_width=True):
            with st.spinner("Autenticando..."):
                r = req("POST", "/login", data={"email": login_email, "password": login_pass})
            if r.ok:
                data = r.json()
                token = data.get("access_token") or data.get("token", "")
                if token:
                    st.session_state.auth_token = token
                    st.success("Login realizado com sucesso!")
                else:
                    st.error("Token não encontrado na resposta.")
                    show_response(r)
            else:
                st.session_state.auth_token = None
                show_response(r)

    if st.session_state.auth_token:
        st.info("✅ Autenticado. Token armazenado na sessão.")
    else:
        st.info("Faça login para acessar os recursos autenticados.")

    # ── Section 2: Bank Account ────────────────────────────────────────────
    st.subheader("2. Conta bancária")
    col1, col2 = st.columns(2)
    with col1:
        ba_name = st.text_input("Nome da conta", placeholder="Conta Principal", key="ba_name")
    with col2:
        ba_bank = st.text_input("Banco", placeholder="Nubank", key="ba_bank")

    col3, col4 = st.columns(2)
    with col3:
        ba_pix_key = st.text_input("Chave PIX (opcional)", placeholder="12345678900", key="ba_pix")
    with col4:
        ba_pix_titular = st.text_input("Titular PIX (opcional)", placeholder="Loja Teste", key="ba_titular")

    if st.button("📋 Criar conta bancária", use_container_width=True):
        if not ba_name or not ba_bank:
            st.error("Preencha nome da conta e banco.")
        elif not st.session_state.auth_token:
            st.error("Faça login primeiro (seção 1).")
        else:
            payload = {"name": ba_name, "bank_name": ba_bank}
            if ba_pix_key.strip():
                payload["pix_key"] = ba_pix_key.strip()
            if ba_pix_titular.strip():
                payload["pix_titular"] = ba_pix_titular.strip()

            with st.spinner("Criando conta bancária..."):
                r = req(
                    "POST",
                    "/estabelecimentos/contasbancarias",
                    data=payload,
                    token=st.session_state.auth_token,
                )
            if r.ok:
                data = r.json()
                conta = data.get("conta", data)
                cid = conta.get("id")
                if cid:
                    st.session_state.bank_account_id = cid
                    st.success(f"Conta bancária criada! ID: `{cid}`")
                else:
                    st.success("Conta bancária criada!")
                    show_response(r)
            else:
                show_response(r)

    if st.button("📋 Listar contas bancárias"):
        if not st.session_state.auth_token:
            st.error("Faça login primeiro (seção 1).")
        else:
            with st.spinner("Listando contas..."):
                r = req("GET", "/estabelecimentos/contasbancarias", token=st.session_state.auth_token)
            if r.ok:
                data = r.json()
                contas = data.get("contas", data if isinstance(data, list) else [])
                if contas:
                    st.dataframe(
                        [
                            {
                                "ID": c.get("id"),
                                "Nome": c.get("name", ""),
                                "Banco": c.get("bank_name", ""),
                                "Chave PIX": c.get("pix_key", ""),
                            }
                            for c in contas
                        ],
                        use_container_width=True,
                    )
                else:
                    st.info("Nenhuma conta bancária encontrada.")
            else:
                show_response(r)

    # ── Section 3: Saque ───────────────────────────────────────────────────
    st.subheader("3. Saque")

    col1, col2 = st.columns(2)
    with col1:
        ba_id = st.number_input(
            "ID da conta bancária",
            min_value=1,
            value=st.session_state.bank_account_id or 1,
            key="saque_ba_id",
        )
    with col2:
        saque_desc = st.text_input("Descrição", placeholder="Saque de teste", key="saque_desc")

    col3, col4 = st.columns(2)
    with col3:
        saque_value = st.text_input("Valor solicitado (R$)", placeholder="250.00", key="saque_val")
    with col4:
        st.write("")
        st.write("")

    if st.button("💸 Solicitar saque", type="primary", use_container_width=True):
        if not saque_desc or not saque_value:
            st.error("Preencha descrição e valor.")
        elif not st.session_state.auth_token:
            st.error("Faça login primeiro (seção 1).")
        else:
            payload = {
                "description": saque_desc,
                "asked_value": saque_value,
                "bank_account": int(ba_id),
            }
            with st.spinner("Criando saque..."):
                r = req(
                    "POST",
                    "/estabelecimentos/saques",
                    data=payload,
                    token=st.session_state.auth_token,
                )
            if r.ok:
                data = r.json()
                saque = data.get("saque", data)
                sid = saque.get("id")
                sstatus = saque.get("status", "?")
                st.success(f"Saque criado! ID: `{sid}`  \nStatus: `{sstatus}`")
                st.balloons()
            else:
                show_response(r)

    if st.button("📋 Listar saques"):
        if not st.session_state.auth_token:
            st.error("Faça login primeiro (seção 1).")
        else:
            with st.spinner("Listando saques..."):
                r = req("GET", "/estabelecimentos/saques", token=st.session_state.auth_token)
            if r.ok:
                data = r.json()
                saques = data.get("saques", data if isinstance(data, list) else [])
                if saques:
                    st.dataframe(
                        [
                            {
                                "ID": s.get("id"),
                                "Descrição": s.get("description", ""),
                                "Valor": s.get("asked_value", ""),
                                "Status": s.get("status", ""),
                            }
                            for s in saques
                        ],
                        use_container_width=True,
                    )
                else:
                    st.info("Nenhum saque encontrado.")
            else:
                show_response(r)

# ─────────────────────────────────────────────────────────────────────────────
# MONITORES
# ─────────────────────────────────────────────────────────────────────────────

elif menu == "Monitores":
    st.header("👤 Monitores")
    st.markdown("CRUD de monitores (consignees) do estabelecimento.")

    # ── Section 1: Login ───────────────────────────────────────────────────
    st.subheader("1. Login (estabelecimento)")
    col1, col2, col3 = st.columns([2, 2, 1])
    with col1:
        mon_email = st.text_input("E-mail", value="", key="mon_email")
    with col2:
        mon_pass = st.text_input("Senha", type="password", value="", key="mon_pass")
    with col3:
        st.write("")
        st.write("")
        if st.button("🔑 Entrar", use_container_width=True, key="mon_login_btn"):
            with st.spinner("Autenticando..."):
                r = req("POST", "/login", data={"email": mon_email, "password": mon_pass})
            if r.ok:
                data = r.json()
                token = data.get("access_token") or data.get("token", "")
                if token:
                    st.session_state.auth_token = token
                    st.success("Login realizado com sucesso!")
                else:
                    st.error("Token não encontrado na resposta.")
                    show_response(r)
            else:
                st.session_state.auth_token = None
                show_response(r)

    if st.session_state.auth_token:
        st.info("✅ Autenticado. Token armazenado na sessão.")
    else:
        st.info("Faça login para acessar os recursos autenticados.")

    # ── Section 2: Create Monitor ──────────────────────────────────────────
    st.subheader("2. Criar monitor")
    col1, col2 = st.columns(2)
    with col1:
        cr_name = st.text_input("Nome", placeholder="João Silva", key="mon_cr_name")
    with col2:
        cr_email = st.text_input("E-mail", placeholder="joao@exemplo.com", key="mon_cr_email")

    col3, col4 = st.columns(2)
    with col3:
        cr_phone = st.text_input("Telefone", placeholder="11999999999", key="mon_cr_phone")
    with col4:
        st.write("")

    with st.expander("Endereço", expanded=True):
        ac1, ac2 = st.columns(2)
        with ac1:
            cr_street = st.text_input("Rua", placeholder="Rua Nova", key="mon_cr_street")
            cr_neigh = st.text_input("Bairro", placeholder="Centro", key="mon_cr_neigh")
            cr_city = st.text_input("Cidade", placeholder="Campinas", key="mon_cr_city")
        with ac2:
            cr_number = st.text_input("Número", placeholder="50", key="mon_cr_number")
            cr_state = st.text_input("Estado", placeholder="SP", key="mon_cr_state")
            cr_cep = st.text_input("CEP", placeholder="13000000", key="mon_cr_cep")

        cr_complement = st.text_input("Complemento", placeholder="Apto 42", key="mon_cr_complement")

    if st.button("👤 Criar monitor", type="primary", use_container_width=True, key="mon_create_btn"):
        if not cr_name or not cr_email or not cr_phone:
            st.error("Preencha nome, e-mail e telefone.")
        elif not st.session_state.auth_token:
            st.error("Faça login primeiro (seção 1).")
        else:
            payload = {
                "name": cr_name,
                "email": cr_email,
                "phone": cr_phone,
                "address": {
                    "street": cr_street,
                    "number": cr_number,
                    "neighborhood": cr_neigh,
                    "city": cr_city,
                    "state": cr_state,
                    "cep": cr_cep,
                    "complement": cr_complement,
                },
            }
            with st.spinner("Criando monitor..."):
                r = req("POST", "/monitores", data=payload, token=st.session_state.auth_token)
            if r.ok:
                data = r.json()
                monitor = data.get("monitor", data)
                mid = monitor.get("id")
                st.success(f"Monitor criado! ID: `{mid}`")
                st.balloons()
            else:
                show_response(r)

    # ── Section 3: List Monitors ───────────────────────────────────────────
    st.subheader("3. Listar monitores")
    if st.button("📋 Listar monitores", use_container_width=True, key="mon_list_btn"):
        if not st.session_state.auth_token:
            st.error("Faça login primeiro (seção 1).")
        else:
            with st.spinner("Listando monitores..."):
                r = req("GET", "/monitores", token=st.session_state.auth_token)
            if r.ok:
                data = r.json()
                monitores = data.get("data", data if isinstance(data, list) else [])
                if monitores:
                    st.dataframe(
                        [
                            {
                                "ID": m.get("id"),
                                "Nome": m.get("name", ""),
                                "E-mail": m.get("email", ""),
                                "Telefone": m.get("phone", ""),
                            }
                            for m in monitores
                        ],
                        use_container_width=True,
                    )
                else:
                    st.info("Nenhum monitor encontrado.")
            else:
                show_response(r)

    # ── Section 4: Update Monitor ──────────────────────────────────────────
    st.subheader("4. Editar monitor")
    st.markdown("Edição parcial — apenas os campos preenchidos são alterados.")

    up_id = st.number_input("ID do monitor", min_value=1, value=1, key="mon_up_id")

    col1, col2 = st.columns(2)
    with col1:
        up_name = st.text_input("Novo nome (opcional)", placeholder="", key="mon_up_name")
    with col2:
        up_email = st.text_input("Novo e-mail (opcional)", placeholder="", key="mon_up_email")

    up_phone = st.text_input("Novo telefone (opcional)", placeholder="", key="mon_up_phone")

    with st.expander("Editar endereço (opcional)", expanded=False):
        ae1, ae2 = st.columns(2)
        with ae1:
            up_street = st.text_input("Rua", placeholder="", key="mon_up_street")
            up_neigh = st.text_input("Bairro", placeholder="", key="mon_up_neigh")
            up_city = st.text_input("Cidade", placeholder="", key="mon_up_city")
        with ae2:
            up_number = st.text_input("Número", placeholder="", key="mon_up_number")
            up_state = st.text_input("Estado", placeholder="", key="mon_up_state")
            up_cep = st.text_input("CEP", placeholder="", key="mon_up_cep")

        up_complement = st.text_input("Complemento", placeholder="", key="mon_up_complement")

    if st.button("💾 Salvar alterações", type="primary", use_container_width=True, key="mon_up_btn"):
        if not st.session_state.auth_token:
            st.error("Faça login primeiro (seção 1).")
        else:
            payload = {}
            if up_name.strip():
                payload["name"] = up_name.strip()
            if up_email.strip():
                payload["email"] = up_email.strip()
            if up_phone.strip():
                payload["phone"] = up_phone.strip()

            addr = {}
            if up_street.strip():
                addr["street"] = up_street.strip()
            if up_number.strip():
                addr["number"] = up_number.strip()
            if up_neigh.strip():
                addr["neighborhood"] = up_neigh.strip()
            if up_city.strip():
                addr["city"] = up_city.strip()
            if up_state.strip():
                addr["state"] = up_state.strip()
            if up_cep.strip():
                addr["cep"] = up_cep.strip()
            if up_complement.strip():
                addr["complement"] = up_complement.strip()

            if addr:
                payload["address"] = addr

            if not payload:
                st.info("Nenhum campo preenchido — PUT com corpo vazio mantém os dados inalterados.")

            with st.spinner("Atualizando monitor..."):
                r = req(
                    "PUT",
                    f"/monitores/{up_id}",
                    data=payload,
                    token=st.session_state.auth_token,
                )
            if r.ok:
                data = r.json()
                monitor = data.get("monitor", data)
                st.success("Monitor atualizado!")
                st.code(json.dumps(monitor, indent=2, ensure_ascii=False), language="json")
            else:
                show_response(r)

    # ── Section 5: Delete Monitor ──────────────────────────────────────────
    st.subheader("5. Excluir monitor")
    del_id = st.number_input("ID do monitor para excluir", min_value=1, value=1, key="mon_del_id")

    if st.button("🗑️ Excluir monitor", use_container_width=True, key="mon_del_btn"):
        if not st.session_state.auth_token:
            st.error("Faça login primeiro (seção 1).")
        else:
            with st.spinner("Excluindo monitor..."):
                r = req("DELETE", f"/monitores/{del_id}", token=st.session_state.auth_token)
            if r.ok:
                st.success(f"Monitor `{del_id}` excluído!")
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
            help="Caminho da API, ex: /produtos, /login, /carrinho/dados",
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

        # Resolve auth
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

        # Build URL
        path = endpoint.strip().lstrip("/")
        params = None
        if query_str.strip():
            parsed = urlparse("?" + query_str.strip())
            params = {k: v[0] if len(v) == 1 else v for k, v in parse_qs(parsed.query).items()}

        # Parse body
        data_payload = None
        if body_visible and body_text.strip():
            try:
                data_payload = json.loads(body_text.strip())
            except json.JSONDecodeError as e:
                st.error(f"JSON inválido no corpo: {e}")
                st.stop()

        # Send
        with st.spinner(f"Enviando {http_method} {path}..."):
            r = req(
                http_method,
                path,
                params=params,
                data=data_payload,
                token=token_override,
            )
        show_response(r)
