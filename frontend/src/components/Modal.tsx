export default function Modal({ children, onClose, width, maxHeight }: {
  children: React.ReactNode
  onClose: () => void
  width?: number
  maxHeight?: string
}) {
  return (
    <div onClick={onClose} style={{ position: 'fixed', inset: 0, zIndex: 100, background: 'rgba(0,0,0,0.5)', backdropFilter: 'blur(4px)', display: 'flex', alignItems: 'center', justifyContent: 'center', animation: 'fadeIn 0.2s ease' }}>
      <div onClick={e => e.stopPropagation()} style={{ position: 'relative', background: 'var(--card)', border: '1px solid var(--border)', borderRadius: 'var(--radius)', width: width ?? 700, maxWidth: '90vw', maxHeight: maxHeight ?? '85vh', overflowY: 'auto', padding: 28, boxShadow: '0 20px 60px rgba(0,0,0,0.3)', animation: 'slideUp 0.25s ease' }}>
        <button onClick={onClose} style={{ position: 'absolute', top: 14, right: 14, width: 30, height: 30, borderRadius: '50%', border: '1px solid var(--border)', background: 'var(--card)', color: 'var(--text-secondary)', fontSize: 16, cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>✕</button>
        {children}
      </div>
    </div>
  )
}
