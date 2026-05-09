# -*- coding: utf-8 -*-
"""Replace hardcoded English API error fallbacks with t('errors.api.*') in Vue pages."""
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1] / "pages"

REPL = [
    ("|| 'Failed to fetch order details'", "|| t('errors.api.load_failed')"),
    ("|| 'Please confirm manual compliance review before proceeding.'", "|| t('customer.orders.error_compliance_ack')"),
    ("|| 'Failed to confirm order'", "|| t('errors.api.confirm_failed')"),
    ("|| 'Failed to cancel order'", "|| t('errors.api.cancel_failed')"),
    ("|| 'Failed to upload payment proof'", "|| t('errors.api.upload_failed')"),
    ("|| 'Failed to fetch products'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to save product'", "|| t('errors.api.save_failed')"),
    ("|| 'Failed to delete product'", "|| t('errors.api.delete_failed')"),
    ("|| 'Failed to fetch orders'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to load order details'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to save order'", "|| t('errors.api.save_failed')"),
    ("|| 'Failed to delete order'", "|| t('errors.api.delete_failed')"),
    ("|| 'Failed to fetch inquiries'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to load inquiry details'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to save inquiry'", "|| t('errors.api.save_failed')"),
    ("|| 'Failed to delete inquiry'", "|| t('errors.api.delete_failed')"),
    ("|| 'Failed to assign inquiry'", "|| t('errors.api.assign_failed')"),
    ("|| 'Failed to analyze inquiry'", "|| t('errors.api.analyze_failed')"),
    ("|| 'Failed to submit quote'", "|| t('errors.api.quote_failed')"),
    ("|| 'Failed to load product'", "|| t('errors.api.product_load_failed')"),
    ("|| 'Failed to load dashboard data'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to fetch settings'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to save setting'", "|| t('errors.api.save_failed')"),
    ("|| 'Failed to fetch audit log'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to fetch staff'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to fetch activity'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to fetch invoices'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to save invoice'", "|| t('errors.api.save_failed')"),
    ("|| 'Failed to send invoice'", "|| t('errors.api.send_failed')"),
    ("|| 'Failed to load trade details'", "|| t('errors.api.trade_load_failed')"),
    ("|| 'Failed to generate AI draft order'", "|| t('errors.api.ai_draft_failed')"),
    ("|| 'Failed to confirm draft order'", "|| t('errors.api.ai_confirm_failed')"),
    ("|| 'Failed to create manual order'", "|| t('errors.api.manual_order_failed')"),
    ("|| 'Failed to fetch trade'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to update status'", "|| t('errors.api.status_failed')"),
    ("|| 'Failed to confirm document'", "|| t('errors.api.document_confirm_failed')"),
    ("|| 'Failed to fetch trades'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to load company'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to upload document'", "|| t('errors.api.upload_failed')"),
    ("|| 'Failed to load notifications'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to load pricing'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to load quotes'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to save price list'", "|| t('errors.api.save_failed')"),
    ("|| 'Failed to delete price list'", "|| t('errors.api.delete_failed')"),
    ("|| 'Failed to save pricing rule'", "|| t('errors.api.save_failed')"),
    ("|| 'Failed to delete price'", "|| t('errors.api.delete_failed')"),
    ("|| 'Failed to fetch users'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to create user'", "|| t('errors.api.user_create_failed')"),
    ("|| 'Failed to delete user'", "|| t('errors.api.delete_failed')"),
    ("|| 'Failed to fetch inventory'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to adjust stock'", "|| t('errors.api.stock_adjust_failed')"),
    ("|| 'Failed to fetch shipments'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to save shipment'", "|| t('errors.api.save_failed')"),
    ("|| 'Failed to fetch projects'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to fetch order'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to fetch payments'", "|| t('errors.api.fetch_payments_failed')"),
    ("|| 'Failed to confirm payment'", "|| t('errors.api.payment_confirm_failed')"),
    ("|| 'Failed to refund payment'", "|| t('errors.api.refund_failed')"),
    ("|| 'Failed to create trade'", "|| t('errors.api.trade_create_failed')"),
    ("|| 'Failed to fetch inquiry details'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to start trade'", "|| t('errors.api.trade_start_failed')"),
    ("|| 'Failed to create project'", "|| t('errors.api.save_failed')"),
    ("|| 'Failed to fetch project'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to load invoices'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to fetch companies'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to verify company'", "|| t('errors.api.company_verify_failed')"),
    ("|| 'Failed to fetch inquiry'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to convert inquiry'", "|| t('errors.api.inquiry_convert_failed')"),
    ("|| 'Failed to fetch company'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to fetch user'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to update profile'", "|| t('errors.api.save_failed')"),
    ("|| 'Failed to update role'", "|| t('errors.api.save_failed')"),
    ("|| 'Failed to fetch content'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to load content details'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to save content'", "|| t('errors.api.save_failed')"),
    ("|| 'Failed to delete content'", "|| t('errors.api.delete_failed')"),
    ("|| 'Failed to fetch certifications'", "|| t('errors.api.load_failed')"),
    ("|| 'Failed to save certification'", "|| t('errors.api.save_failed')"),
    ("|| 'Failed to delete certification'", "|| t('errors.api.delete_failed')"),
    ("err.message || 'Failed to fetch user'", "err.message || t('errors.api.load_failed')"),
]


def main():
    for path in sorted(ROOT.rglob("*.vue")):
        text = path.read_text(encoding="utf-8")
        orig = text
        for a, b in REPL:
            text = text.replace(a, b)
        text = text.replace(
            "richDocError.value = err?.message || 'Failed to save'",
            "richDocError.value = err?.message || t('errors.api.document_save_failed')",
        )
        text = text.replace("e?.message || 'Failed to load company'", "e?.message || t('errors.api.load_failed')")
        text = text.replace("e?.message || 'Failed to save'", "e?.message || t('errors.api.save_failed')")
        text = text.replace("e?.message || 'Failed to upload document'", "e?.message || t('errors.api.upload_failed')")
        text = text.replace("e?.message || 'Failed to load invoices'", "e?.message || t('errors.api.load_failed')")
        text = text.replace("e?.message || 'Failed to load pricing'", "e?.message || t('errors.api.load_failed')")
        text = text.replace("|| 'Upload failed'", "|| t('errors.api.upload_failed')")
        if text != orig:
            path.write_text(text, encoding="utf-8")
            print("updated", path.relative_to(ROOT.parent))


if __name__ == "__main__":
    main()
