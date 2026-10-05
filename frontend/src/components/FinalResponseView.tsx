type Props = { finalDoc: string; onDownload: ()=>void; onDelete: ()=>void; };

export default function FinalResponseView({ finalDoc, onDownload, onDelete }: Props) {
  return (
    <div style={{ padding: 16, border: '1px solid #ddd', borderRadius: 12, background: '#fafafa' }}>
      <div style={{ display: 'flex', gap: 12, marginBottom: 12 }}>
        <button onClick={onDownload} style={{ background: '#0a7', color: '#fff', padding: '8px 16px', borderRadius: 8 }}>Download Final Response</button>
        <button onClick={()=> { if(confirm('Sab delete ho jayega, sure?')) onDelete(); }} style={{ background: '#c00', color: '#fff', padding: '8px 16px', borderRadius: 8 }}>Delete Everything</button>
      </div>
      <pre style={{ whiteSpace: 'pre-wrap', fontFamily: 'inherit', fontSize: 13, lineHeight: 1.6 }}>{finalDoc}</pre>
    </div>
  );
}
