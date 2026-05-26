#!/usr/bin/env node
/**
 * Apply locale-specific customer dashboard/nav translations over English fallbacks.
 */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.join(__dirname, '..', 'i18n')

/** @type {Record<string, Record<string, unknown>>} */
const patches = {
  ko: {
    customer: {
      nav: { returns: '반품 신청' },
      a11y: {
        collapseSidebar: '사이드바 접기',
        expandSidebar: '사이드바 펼치기',
      },
      page_titles: { returns: '반품 신청' },
    },
  },
  ja: {
    customer: {
      brand: 'CandyPro OEM',
      logout: 'ログアウト',
      nav: {
        dashboard: 'ダッシュボード',
        products: '製品',
        logout: 'ログアウト',
        cart: '注文リクエスト',
        orders: '注文一覧',
        inquiries: 'お問い合わせ',
        oemProjects: 'OEMプロジェクト',
        quotes: '見積一覧',
        notifications: '通知',
        trades: '取引一覧',
        invoices: '請求書',
        shipments: '配送追跡',
        returns: '返品申請',
        company: '会社プロフィール',
        pricing: '価格設定',
        profile: 'プロフィール',
        help: 'ヘルプ',
        resources: 'リソース',
      },
      a11y: {
        openMenu: 'ナビゲーションメニューを開く',
        closeMenu: 'ナビゲーションメニューを閉じる',
        collapseSidebar: 'サイドバーを折りたたむ',
        expandSidebar: 'サイドバーを展開',
        cart: '注文リクエストとカート',
        portalNav: '顧客ポータル',
      },
      page_titles: {
        dashboard: 'ダッシュボード',
        orders: '注文一覧',
        inquiries: 'お問い合わせ',
        quotes: '見積一覧',
        notifications: '通知',
        trades: '取引一覧',
        invoices: '請求書',
        shipments: '配送追跡',
        returns: '返品申請',
        company: '会社プロフィール',
        pricing: '価格設定',
        profile: 'アカウント',
        help: 'ヘルプセンター',
        resources: 'リソース',
      },
      dashboard: {
        loading: 'ダッシュボードを読み込み中...',
        error: 'ダッシュボードの読み込みに失敗しました',
        welcome: 'おかえりなさい、{name} さん！',
        subtitle: 'ポータルから注文、配送追跡、B2Bプロフィールを管理できます。',
        recent_orders: '最近の注文',
        orders_count: '注文 {count} 件',
        view_all_orders: 'すべての注文を見る',
        company_info: '会社情報',
        no_company: '登録された会社がありません',
        update_profile: 'プロフィールを更新',
        latest_orders: '最新の注文',
        placed_on: '注文日',
        no_orders: 'まだ注文がありません。',
        no_orders_desc: '製品を閲覧して最初の注文を開始してください。',
      },
      pending: {
        banner: 'アカウントは承認待ちです。閲覧は可能ですが、承認後に注文できます。',
        dashboard_title: 'アカウント承認待ち',
        dashboard_message: 'B2Bアカウントを審査中です。承認後に注文できるようになります。',
        step_browse: '製品を閲覧',
        step_review: '審査中',
        step_order: '注文開始',
      },
    },
  },
  vi: {
    customer: {
      nav: {
        dashboard: 'Bảng điều khiển',
        products: 'Sản phẩm',
        logout: 'Đăng xuất',
        cart: 'Yêu cầu đặt hàng',
        orders: 'Đơn hàng của tôi',
        inquiries: 'Yêu cầu báo giá',
        oemProjects: 'Dự án OEM',
        quotes: 'Báo giá của tôi',
        notifications: 'Thông báo',
        trades: 'Giao dịch của tôi',
        invoices: 'Hóa đơn',
        shipments: 'Theo dõi vận chuyển',
        returns: 'Trả hàng',
        company: 'Hồ sơ công ty',
        pricing: 'Bảng giá của tôi',
        profile: 'Hồ sơ',
        help: 'Trợ giúp',
        resources: 'Tài nguyên',
      },
      a11y: {
        openMenu: 'Mở menu điều hướng',
        closeMenu: 'Đóng menu điều hướng',
        collapseSidebar: 'Thu gọn thanh bên',
        expandSidebar: 'Mở rộng thanh bên',
        cart: 'Yêu cầu đặt hàng và giỏ hàng',
        portalNav: 'Cổng khách hàng',
      },
      page_titles: {
        dashboard: 'Bảng điều khiển',
        orders: 'Đơn hàng của tôi',
        inquiries: 'Yêu cầu báo giá',
        quotes: 'Báo giá của tôi',
        notifications: 'Thông báo',
        trades: 'Giao dịch của tôi',
        invoices: 'Hóa đơn',
        shipments: 'Theo dõi vận chuyển',
        returns: 'Trả hàng',
        company: 'Hồ sơ công ty',
        pricing: 'Bảng giá của tôi',
        profile: 'Hồ sơ tài khoản',
        help: 'Trung tâm trợ giúp',
        resources: 'Tài nguyên',
      },
      dashboard: {
        loading: 'Đang tải bảng điều khiển...',
        error: 'Không thể tải bảng điều khiển',
        welcome: 'Chào mừng trở lại, {name}!',
        subtitle: 'Quản lý đơn hàng, theo dõi vận chuyển và hồ sơ B2B từ cổng khách hàng.',
        recent_orders: 'Đơn hàng gần đây',
        orders_count: '{count} đơn hàng',
        view_all_orders: 'Xem tất cả đơn hàng',
        company_info: 'Thông tin công ty',
        no_company: 'Chưa có công ty',
        update_profile: 'Cập nhật hồ sơ',
        latest_orders: 'Đơn hàng mới nhất',
        placed_on: 'Ngày đặt',
        no_orders: 'Bạn chưa có đơn hàng nào.',
        no_orders_desc: 'Duyệt sản phẩm và đặt đơn hàng đầu tiên.',
      },
      pending: {
        banner: 'Tài khoản đang chờ phê duyệt. Bạn có thể duyệt sản phẩm; đặt hàng sẽ mở sau khi được duyệt.',
        dashboard_title: 'Tài khoản chờ phê duyệt',
        dashboard_message: 'Đội ngũ đang xem xét tài khoản B2B của bạn. Bạn có thể đặt hàng sau khi được duyệt.',
        step_browse: 'Duyệt sản phẩm',
        step_review: 'Đang xem xét',
        step_order: 'Bắt đầu đặt hàng',
      },
    },
  },
  th: {
    customer: {
      nav: {
        dashboard: 'แดชบอร์ด',
        products: 'สินค้า',
        logout: 'ออกจากระบบ',
        cart: 'คำขอสั่งซื้อ',
        orders: 'คำสั่งซื้อของฉัน',
        inquiries: 'คำถามของฉัน',
        oemProjects: 'โปรเจกต์ OEM',
        quotes: 'ใบเสนอราคา',
        notifications: 'การแจ้งเตือน',
        trades: 'การค้าของฉัน',
        invoices: 'ใบแจ้งหนี้',
        shipments: 'ติดตามการจัดส่ง',
        returns: 'คืนสินค้า',
        company: 'โปรไฟล์บริษัท',
        pricing: 'ราคาของฉัน',
        profile: 'โปรไฟล์',
        help: 'ความช่วยเหลือ',
        resources: 'ทรัพยากร',
      },
      a11y: {
        openMenu: 'เปิดเมนูนำทาง',
        closeMenu: 'ปิดเมนูนำทาง',
        collapseSidebar: 'ยุบแถบด้านข้าง',
        expandSidebar: 'ขยายแถบด้านข้าง',
        cart: 'คำขอสั่งซื้อและตะกร้า',
        portalNav: 'พอร์ทัลลูกค้า',
      },
      page_titles: {
        dashboard: 'แดชบอร์ด',
        orders: 'คำสั่งซื้อของฉัน',
        returns: 'คืนสินค้า',
      },
      dashboard: {
        loading: 'กำลังโหลดแดชบอร์ด...',
        error: 'โหลดแดชบอร์ดไม่สำเร็จ',
        welcome: 'ยินดีต้อนรับกลับ, {name}!',
        subtitle: 'จัดการคำสั่งซื้อ การติดตาม และโปรไฟล์ B2B จากพอร์ทัลของคุณ',
        recent_orders: 'คำสั่งซื้อล่าสุด',
        orders_count: '{count} คำสั่งซื้อ',
        view_all_orders: 'ดูคำสั่งซื้อทั้งหมด',
        company_info: 'ข้อมูลบริษัท',
        no_company: 'ไม่มีบริษัทที่ลงทะเบียน',
        update_profile: 'อัปเดตโปรไฟล์',
        latest_orders: 'คำสั่งซื้อล่าสุด',
        placed_on: 'สั่งเมื่อ',
        no_orders: 'คุณยังไม่มีคำสั่งซื้อ',
        no_orders_desc: 'เรียกดูสินค้าและสั่งซื้อครั้งแรกของคุณ',
      },
    },
  },
  id: {
    customer: {
      nav: { returns: 'Pengembalian' },
      page_titles: { returns: 'Pengembalian' },
      a11y: {
        collapseSidebar: 'Ciutkan sidebar',
        expandSidebar: 'Perluas sidebar',
      },
    },
  },
  ms: {
    customer: {
      nav: { returns: 'Pulangan' },
      page_titles: { returns: 'Pulangan' },
      a11y: {
        collapseSidebar: 'Runtuhkan bar sisi',
        expandSidebar: 'Kembangkan bar sisi',
      },
    },
  },
  ar: {
    customer: {
      nav: { returns: 'المرتجعات' },
      a11y: {
        collapseSidebar: 'طي الشريط الجانبي',
        expandSidebar: 'توسيع الشريط الجانبي',
      },
      page_titles: { returns: 'المرتجعات' },
    },
  },
}

function deepAssign(base, patch) {
  for (const [key, value] of Object.entries(patch)) {
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      if (!base[key] || typeof base[key] !== 'object') base[key] = {}
      deepAssign(base[key], value)
    } else {
      base[key] = value
    }
  }
}

for (const [locale, patch] of Object.entries(patches)) {
  const file = path.join(root, locale, 'customer.json')
  const data = JSON.parse(fs.readFileSync(file, 'utf8'))
  deepAssign(data, patch)
  fs.writeFileSync(file, JSON.stringify(data, null, 2) + '\n', 'utf8')
  console.log(`Patched ${locale}/customer.json`)
}
