Use a skill `brainstorming` do Superpowers antes de qualquer trabalho de implementação.

Quero desenvolver uma aplicação auxiliar para o iRacing que permita gerenciar bindings de botões e controles dos carros a partir dos arquivos de configuração locais do simulador.

**Nesta sessão, quero fazer descoberta de produto e discutir os requisitos. Não quero começar a programar.**

## Contexto do problema

O iRacing permite configurar controles e keybinds em um perfil baseline, além de manter configurações específicas para cada carro. Isso permite ter valores específicos, como força de force feedback, e bindings diferentes entre carros, mas torna trabalhoso modificar uma única tecla ou botão em todos os perfis.

Por exemplo, se eu quiser atribuir uma função ao botão PL do volante, preciso alterar manualmente o binding nos perfis dos carros relevantes. Usar somente o baseline também não resolve todos os casos, pois alguns carros precisam manter configurações específicas.

Quero uma aplicação que leia os arquivos locais e permita gerenciar bindings de forma geral e específica por carro, preservando as diferenças que devem continuar existindo.

Os arquivos de interesse incluem:

- Configuração baseline: `/mnt/c/Users/Ian/Documents/iRacing/profiles/controls/Baseline control/controls.cfg` e `joyCalib.yaml`, se presentes nesse perfil.
- Configurações específicas por carro, por exemplo, `/mnt/c/Users/Ian/Documents/iRacing/profiles/controls/audigto/controls.cfg` e `/mnt/c/Users/Ian/Documents/iRacing/profiles/controls/f4/controls.cfg`.
- Os arquivos `joyCalib.yaml` correspondentes, quando presentes.

Esses caminhos são exemplos do meu ambiente. A solução não deve depender de um nome de usuário fixo, e o diretório de dados deverá ser configurável ou detectado de forma confiável.

## Como quero trabalhar

1. Conduza uma conversa iterativa em português, fazendo uma pergunta por resposta.
2. Comece pelas dúvidas de produto e comportamento que tenham maior impacto. Não apresente uma lista extensa de perguntas.
3. Ajude-me a distinguir requisitos confirmados, hipóteses e questões em aberto.
4. Discuta alternativas de comportamento e suas consequências antes de escolher uma abordagem técnica.
5. Não assuma como o iRacing combina o baseline e os arquivos específicos por carro. Esse comportamento precisa ser investigado e validado.
6. Diferencie configurações de bindings de teclado e botões de calibração de eixos e dispositivos. Não presuma que `controls.cfg` e `joyCalib.yaml` tenham o mesmo propósito ou devam receber as mesmas alterações.
7. Investigue as estruturas reais dos arquivos somente depois de discutirmos como fazê-lo, usando exemplos representativos e leitura sem escrita. Não invente o formato dos arquivos.
8. Não escolha a stack tecnológica antes de entendermos os requisitos de execução, interface, distribuição e acesso aos arquivos.
9. Não crie arquivos, não gere código, não instale dependências, não execute scaffolding e não altere nenhuma configuração do iRacing durante esta fase.
10. Quando houver informação suficiente, apresente uma proposta de produto com alternativas e trade-offs. Peça minha aprovação explícita antes de registrar uma especificação final ou iniciar o planejamento de implementação.

## Resultado esperado desta fase

Quero sair da conversa com uma definição clara do problema, do comportamento esperado da aplicação, das regras de edição geral e por carro, dos riscos relacionados aos arquivos de configuração, do escopo de um primeiro protótipo e das decisões técnicas que ainda precisaremos tomar.

Comece seguindo a skill `brainstorming` e faça uma única pergunta aberta para iniciar a conversa. Não escreva código nem modifique arquivos.