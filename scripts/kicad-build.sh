#!/bin/sh
echo kicad-builder

[ -d ${BUILD_DIR}/dist ] || mkdir ${BUILD_DIR}/dist

find -name '*.sch' -exec sed -i "s/\$PCBVER/${BUILD_NUMBER}/g" {} \;
find -name '*.net' -exec sed -i "s/\$PCBVER/${BUILD_NUMBER}/g" {} \;
find -name '*.kicad_pcb' -exec sed -i "s/\$PCBVER/${BUILD_NUMBER}/g" {} \;

if [ -d manufacturing ]; then
zip -r dist/manufacturing.zip manufacturing/
fi

cd kicad-files 
for MODULE in `ls -d *`; do
    PCBS=`cd $MODULE && find -name '*.kicad_pcb'`
    for pcb in ${PCBS} ; do
	echo building $MODULE $pcb
	OUTDIR=${BUILD_DIR}/build/${MODULE}
	mkdir ${OUTDIR}
	if [ ! -d ${OUTDIR} ]; then
	    echo "failed to create ${OUTDIR}"
	    exit 10
	fi
	${SCRIPTDIR}/plot_board.py ${MODULE}/$pcb ${OUTDIR}
	( cd ${BUILD_DIR}/build && tar -jcvf ${BUILD_DIR}/dist/kicad_${MODULE}-files.tar.bz2 ${MODULE} ) || exit 10
    done
done

