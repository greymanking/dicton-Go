from os.path import exists

mp3dir="d:/GoSpace/src/greyman/dicton/mp3s/"
wordsdir="d:/words.txt"

fexist=open("d:/exist.txt","w",encoding="utf-8")
fnone=open("d:/none.txt","w",encoding="utf-8")

words=open(wordsdir,"r",encoding="utf-8")

for l in words:
    word,exp=l.split("\t")
    mp3=mp3dir+word.replace(" ","_")+".mp3"
    if exists(mp3):
        fexist.write("%s\t%s\t%s\n"%(word,exp.rstrip(),mp3))
    else:
        fnone.write(word+"\n")
        
fexist.close()
fnone.close()
words.close()
