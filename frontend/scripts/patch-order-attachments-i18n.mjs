// patch-order-attachments-i18n.mjs — 订单消息附件与付款凭证 i18n
import fs from 'fs'
import path from 'path'

const root = path.join(import.meta.dirname, '..', 'i18n')
const locales = ['en', 'zh', 'ko', 'th', 'ja', 'vi', 'id', 'ms', 'ar']

const keys = {
  en: {
    add_attachment: 'Attach files',
    message_attachments_hint: 'Attachments: PDF, DOC, DOCX, PNG, JPG, JPEG, MP4 — max 64MB each, up to 3 files per message.',
    message_max_files: 'Maximum {count} files per message.',
    message_file_type_invalid: 'File type not allowed. Accepted: PDF, DOC, DOCX, PNG, JPG, JPEG, MP4.',
    payment_proof_hint: 'Upload a clear bank transfer receipt or SWIFT confirmation. Accepted formats: PDF, JPG, PNG — max 10MB. Include payer name, amount, date, and reference number.',
    payment_proof_type_invalid: 'Invalid file type. Please upload PDF, JPG, or PNG.',
    payment_proof_too_large: 'File exceeds {max}MB limit.',
  },
  zh: {
    add_attachment: '添加附件',
    message_attachments_hint: '支持 PDF、DOC、DOCX、PNG、JPG、JPEG、MP4，单文件最大 64MB，每条消息最多 3 个附件。',
    message_max_files: '每条消息最多 {count} 个附件。',
    message_file_type_invalid: '不支持的文件类型。允许：PDF、DOC、DOCX、PNG、JPG、JPEG、MP4。',
    payment_proof_hint: '请上传清晰的银行转账回单或 SWIFT 凭证。格式：PDF、JPG、PNG，最大 10MB。须包含付款人、金额、日期及参考号/流水号。',
    payment_proof_type_invalid: '文件格式不正确，请上传 PDF、JPG 或 PNG。',
    payment_proof_too_large: '文件超过 {max}MB 限制。',
  },
  ko: {
    add_attachment: '파일 첨부',
    message_attachments_hint: 'PDF, DOC, DOCX, PNG, JPG, JPEG, MP4 — 파일당 최대 64MB, 메시지당 최대 3개.',
    message_max_files: '메시지당 최대 {count}개 파일.',
    message_file_type_invalid: '허용되지 않는 파일 형식. PDF, DOC, DOCX, PNG, JPG, JPEG, MP4만 가능.',
    payment_proof_hint: '명확한 은행 이체 영수증 또는 SWIFT 확인서를 업로드하세요. PDF, JPG, PNG — 최대 10MB. 송금인, 금액, 날짜, 참조번호 포함.',
    payment_proof_type_invalid: 'PDF, JPG 또는 PNG만 업로드 가능합니다.',
    payment_proof_too_large: '파일이 {max}MB 제한을 초과합니다.',
  },
  th: {
    add_attachment: 'แนบไฟล์',
    message_attachments_hint: 'รองรับ PDF, DOC, DOCX, PNG, JPG, JPEG, MP4 — สูงสุด 64MB/ไฟล์, 3 ไฟล์/ข้อความ',
    message_max_files: 'สูงสุด {count} ไฟล์ต่อข้อความ',
    message_file_type_invalid: 'ประเภทไฟล์ไม่รองรับ: PDF, DOC, DOCX, PNG, JPG, JPEG, MP4',
    payment_proof_hint: 'อัปโหลดใบเสร็จโอนเงินหรือ SWIFT ที่ชัดเจน รองรับ PDF, JPG, PNG สูงสุด 10MB ต้องมีชื่อผู้โอน จำนวนเงิน วันที่ และเลขอ้างอิง',
    payment_proof_type_invalid: 'อัปโหลดได้เฉพาะ PDF, JPG หรือ PNG',
    payment_proof_too_large: 'ไฟล์เกิน {max}MB',
  },
  ja: {
    add_attachment: 'ファイルを添付',
    message_attachments_hint: 'PDF、DOC、DOCX、PNG、JPG、JPEG、MP4 — 1ファイル最大64MB、1メッセージ最大3ファイル。',
    message_max_files: '1メッセージあたり最大{count}ファイル。',
    message_file_type_invalid: '許可されていない形式。PDF、DOC、DOCX、PNG、JPG、JPEG、MP4のみ。',
    payment_proof_hint: '銀行振込明細またはSWIFT確認書をアップロード。PDF/JPG/PNG、最大10MB。振込人・金額・日付・参照番号を含めてください。',
    payment_proof_type_invalid: 'PDF、JPG、PNGのみアップロード可能です。',
    payment_proof_too_large: 'ファイルが{max}MBを超えています。',
  },
  vi: {
    add_attachment: 'Đính kèm tệp',
    message_attachments_hint: 'PDF, DOC, DOCX, PNG, JPG, JPEG, MP4 — tối đa 64MB/tệp, 3 tệp/tin nhắn.',
    message_max_files: 'Tối đa {count} tệp mỗi tin nhắn.',
    message_file_type_invalid: 'Loại tệp không được phép. Chấp nhận: PDF, DOC, DOCX, PNG, JPG, JPEG, MP4.',
    payment_proof_hint: 'Tải lên biên lai chuyển khoản hoặc xác nhận SWIFT rõ ràng. PDF, JPG, PNG — tối đa 10MB. Bao gồm tên người trả, số tiền, ngày và mã tham chiếu.',
    payment_proof_type_invalid: 'Chỉ chấp nhận PDF, JPG hoặc PNG.',
    payment_proof_too_large: 'Tệp vượt quá {max}MB.',
  },
  id: {
    add_attachment: 'Lampirkan file',
    message_attachments_hint: 'PDF, DOC, DOCX, PNG, JPG, JPEG, MP4 — maks. 64MB/file, 3 file/pesan.',
    message_max_files: 'Maksimal {count} file per pesan.',
    message_file_type_invalid: 'Jenis file tidak diizinkan. Diterima: PDF, DOC, DOCX, PNG, JPG, JPEG, MP4.',
    payment_proof_hint: 'Unggah bukti transfer bank atau konfirmasi SWIFT yang jelas. PDF, JPG, PNG — maks. 10MB. Sertakan nama pembayar, jumlah, tanggal, dan nomor referensi.',
    payment_proof_type_invalid: 'Hanya PDF, JPG, atau PNG yang diizinkan.',
    payment_proof_too_large: 'File melebihi batas {max}MB.',
  },
  ms: {
    add_attachment: 'Lampirkan fail',
    message_attachments_hint: 'PDF, DOC, DOCX, PNG, JPG, JPEG, MP4 — maks. 64MB/fail, 3 fail/mesej.',
    message_max_files: 'Maksimum {count} fail setiap mesej.',
    message_file_type_invalid: 'Jenis fail tidak dibenarkan. Diterima: PDF, DOC, DOCX, PNG, JPG, JPEG, MP4.',
    payment_proof_hint: 'Muat naik resit pindahan bank atau pengesahan SWIFT yang jelas. PDF, JPG, PNG — maks. 10MB. Sertakan nama pembayar, jumlah, tarikh, dan nombor rujukan.',
    payment_proof_type_invalid: 'Hanya PDF, JPG atau PNG dibenarkan.',
    payment_proof_too_large: 'Fail melebihi had {max}MB.',
  },
  ar: {
    add_attachment: 'إرفاق ملفات',
    message_attachments_hint: 'PDF وDOC وDOCX وPNG وJPG وJPEG وMP4 — حتى 64MB لكل ملف، 3 ملفات لكل رسالة.',
    message_max_files: 'الحد الأقصى {count} ملفات لكل رسالة.',
    message_file_type_invalid: 'نوع الملف غير مسموح. المقبول: PDF وDOC وDOCX وPNG وJPG وJPEG وMP4.',
    payment_proof_hint: 'ارفع إيصال تحويل بنكي أو تأكيد SWIFT واضح. PDF أو JPG أو PNG — حتى 10MB. يجب أن يتضمن اسم الدافع والمبلغ والتاريخ والرقم المرجعي.',
    payment_proof_type_invalid: 'يُسمح فقط بـ PDF أو JPG أو PNG.',
    payment_proof_too_large: 'الملف يتجاوز حد {max}MB.',
  },
}

for (const loc of locales) {
  const file = path.join(root, loc, 'customer.json')
  const data = JSON.parse(fs.readFileSync(file, 'utf8'))
  if (!data.customer?.orders) {
    console.warn('skip', loc)
    continue
  }
  Object.assign(data.customer.orders, keys[loc])
  fs.writeFileSync(file, JSON.stringify(data, null, 2) + '\n', 'utf8')
  console.log('patched', loc)
}

console.log('done')
