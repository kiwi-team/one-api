import { marked } from 'marked';
import React, { useEffect, useRef, useState } from 'react';
import { Header, Segment } from 'semantic-ui-react';
import { API, showError } from '../../helpers';

const About = () => {
  const [about, setAbout] = useState('');
  const [aboutLoaded, setAboutLoaded] = useState(false);
  const [streamContent, setStreamContent] = useState('');
  const [loading, setLoading] = useState(false);

  // 解析流式数据
  const fetchStreamContent = async () => {
    const requestBody = {
      model: 'deepseek-chat',
      stream: true,
      messages: [
        {
          role: 'user',
          content: '1+2+3+4+..+10=?  并且解释一下勾股定理',
        },
      ],
    };

    setLoading(true);
    try {
      const response = await fetch('http://127.0.0.1:3000/v1/chat/completions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization':'sk-8zXmLSNZvtfLgNfDE22c0f789d1543A0Bf95A2Da9e56AaCf',
        },
        body: JSON.stringify(requestBody),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder('utf-8');
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });

        // 解析流数据
        const lines = buffer.split('\n');
        buffer = lines.pop(); // 保留不完整的数据
        for (const line of lines) {
          if (line.startsWith('data:')) {
            const data = line.slice(5).trim();
            if (data) {
                if(data == '[DONE]') {
                    // 结束
                }else if(data.startsWith('[BAD]')){
                    // 回答没有通过安全审查
                }else {
                    const jsonData = JSON.parse(data);
                    if (jsonData.choices && jsonData.choices[0] && jsonData.choices[0].delta) {
                        const content = jsonData.choices[0].delta.content;
                        if (content) {
                        setStreamContent((prev) => prev + content); // 动态更新内容
                        }
                    }
                }
            }
          }
        }
      }
    } catch (error) {
      showError('加载流内容失败: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  const displayAbout = async () => {
    setAbout(localStorage.getItem('about') || '');
    const res = await API.get('/api/about');
    const { success, message, data } = res.data;
    if (success) {
      let aboutContent = data;
      if (!data.startsWith('https://')) {
        aboutContent = marked.parse(data);
      }
      setAbout(aboutContent);
      localStorage.setItem('about', aboutContent);
    } else {
      showError(message);
      setAbout('加载关于内容失败...');
    }
    setAboutLoaded(true);
  };

  const firstRef = useRef(null)
  useEffect(() => {
    displayAbout().then();
    console.log('122')
    if (!firstRef.current) {
        fetchStreamContent(); // 调用流式接口
        firstRef.current = 1;
    }
  }, []);

  return (
    <>
      {aboutLoaded && about === '' ? (
        <Segment>
          <Header as="h3">关于</Header>
          <p>可在设置页面设置关于内容，支持 HTML & Markdown</p>
          项目仓库地址：
          <a href="https://github.com/songquanpeng/one-api">
            https://github.com/songquanpeng/one-api
          </a>
        </Segment>
      ) : (
        <>
          {about.startsWith('https://') ? (
            <iframe
              src={about}
              style={{ width: '100%', height: '100vh', border: 'none' }}
            />
          ) : (
            <div style={{ fontSize: 'larger' }} dangerouslySetInnerHTML={{ __html: about }}></div>
          )}
        </>
      )}
      {streamContent && (
        <Segment>
          <Header as="h3">流式内容</Header>
          <div style={{ fontSize: 'larger', whiteSpace: 'pre-wrap' }}>{streamContent}</div>
        </Segment>
      )}
      {loading && <p>加载中...</p>}
    </>
  );
};

export default About;
